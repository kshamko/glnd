package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/kshamko/glnd/internal/datasource"
	"github.com/kshamko/glnd/internal/models"
	"github.com/kshamko/glnd/internal/restapi/operations/fees"
	"github.com/sirupsen/logrus"
)

// Fees struct to define Handle func on.
type Fees struct {
	fr     FeesRepo
	logger *logrus.Entry
}

type FeesRepo interface {
	GetFees(context.Context) ([]datasource.Fee, error)
}

func NewFees(fr FeesRepo, logger *logrus.Entry) *Fees {
	return &Fees{
		fr:     fr,
		logger: logger,
	}
}

// Handle function processes http request, needed by swagger generated code.
func (fs *Fees) Handle(in fees.FeesParams) middleware.Responder {
	ctx := context.Background()
	result, err := fs.fr.GetFees(ctx)

	if err != nil && errors.Is(err, datasource.ErrNotFound) {
		return fees.NewFeesNotFound().WithPayload(
			&models.APIInvalidResponse{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			},
		)
	}

	if err != nil {
		return fees.NewFeesInternalServerError().WithPayload(
			&models.APIInvalidResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			},
		)
	}

	response := []*models.Fee{}
	for _, f := range result {
		response = append(
			response,
			&models.Fee{
				T: f.Date.Unix(),
				V: f.Fee,
			},
		)
	}

	return fees.NewFeesOK().WithPayload(
		models.Fees(response),
	)
}
