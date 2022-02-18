FROM golang:1.17-alpine as build

WORKDIR /app
ENV GO111MODULE=on

RUN apk --no-cache add ca-certificates git gcc g++
COPY . .
RUN go mod download
RUN go test ./...
#RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.40.1
#RUN ./bin/golangci-lint run --enable-all --disable goerr113,cyclop,exhaustivestruct,gci,gofumpt,lll,testpackage,wrapcheck,paralleltest 
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /bin/gw cmd/main.go

FROM busybox
COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /bin/gw /
COPY --from=build /app/assets /assets
ENTRYPOINT ["/gw"]
