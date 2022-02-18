package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/InVisionApp/go-health"
	"github.com/go-openapi/loads"
	flags "github.com/jessevdk/go-flags"
	"github.com/kshamko/glnd/internal/datasource"
	"github.com/kshamko/glnd/internal/debug"
	"github.com/kshamko/glnd/internal/handler"
	"github.com/kshamko/glnd/internal/restapi"
	"github.com/kshamko/glnd/internal/restapi/operations"
	"golang.org/x/sync/errgroup"
)

const AppID = "glndapi"

func main() { //nolint: funlen
	var opts = struct {
		HTTPListenHost string `long:"http.listen.host" env:"HTTP_LISTEN_HOST" default:"" description:"http server interface host"`
		HTTPListenPort int    `long:"http.listen.port" env:"HTTP_PORT" default:"8080" description:"http server interface port"`
		DebugListen    string `long:"debug.listen" env:"DEBUG_LISTEN" default:":6060" description:"Interface for serve debug information(metrics/health/pprof)"`
		Verbose        bool   `long:"v" env:"VERBOSE" description:"Enable Verbose log output"`
		PostgresDSN    string `long:"postgres.dsn" env:"POSTGRES_DSN"`
	}{}

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	logger := logrus.WithField("app_id", AppID)
	logger.Logger.SetOutput(os.Stdout)
	logger.Logger.SetLevel(logrus.InfoLevel)

	if opts.Verbose {
		logger.Logger.SetLevel(logrus.DebugLevel)
	}

	logger.Infof("Launching Application with: %+v", opts)

	gr, appctx := errgroup.WithContext(context.Background())
	gr.Go(func() error {
		healthd := health.New()
		d := debug.New(healthd)

		return d.Serve(appctx, opts.DebugListen)
	})

	gr.Go(func() error {
		swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
		if err != nil {
			logger.Error(err)

			return err
		}
		api := operations.NewGLNDAPISwaggerAPI(swaggerSpec)

		feesRepo, err := datasource.NewFeesDS(opts.PostgresDSN)
		if err != nil {
			logger.Error(err)

			return err
		}

		api.FeesFeesHandler = handler.NewFees(feesRepo, logger)

		server := restapi.NewServer(api)

		defer func() {
			_ = server.Shutdown()
		}()

		server.Host = opts.HTTPListenHost
		server.Port = opts.HTTPListenPort

		go func() {
			<-appctx.Done()
			_ = server.Shutdown()
		}()

		return server.Serve()
	})

	errCanceled := errors.New("Canceled")

	gr.Go(func() error {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		cusr := make(chan os.Signal, 1)
		signal.Notify(cusr, syscall.SIGUSR1)
		for {
			select {
			case <-appctx.Done():
				return nil
			case <-sigs:
				logger.Info("Caught stop signal. Exiting ...")

				return errCanceled
			case <-cusr:
				if logger.Level == logrus.DebugLevel {
					logger.Logger.SetLevel(logrus.InfoLevel)
					logger.Info("[INFO] Caught SIGUSR1 signal. Log level changed to INFO")

					continue
				}
				logger.Info("Caught SIGUSR1 signal. Log level changed to DEBUG")
				logger.Logger.SetLevel(logrus.DebugLevel)
			}
		}
	})

	err = gr.Wait()
	if err != nil && errors.Is(err, errCanceled) {
		log.Fatal(err)
	}
}
