# Description

Solution based on https://github.com/kshamko/boilerplate

## Approach

It is important to have API documented, thus there is a problem of documentation and API implementation consistency. To solve the proble the following approach was used:

1. First step is to create swagger spec to describe the API (**api/glnd.swagger.yaml**)
2. Use go-swagger to generate server's code:
```bash
$ make swagger
```
3. Implement endpoints' handlers
4. If there is a requirement to change something in API, start from 1.

## Build and Run

To run the app it would be nice to have **docker-compose** installed

```bash
$ git clone git@github.com:kshamko/glnd.git
$ cd glnd
$ docker-compose up
```

The application will be started on port 8080 and it will be possible to request it localy:
```bash
curl -XGET 'http://localhost:8080/api/v1/fees'
```

or in a browser

## Service Endpoints

Also an additionl port is exposed for debug/healthcheck/swagger-ui purposes

1. Metrics: http://localhost:6060/metrics
2. Healthcheck: http://localhost:6060/healthz
3. Swagger-UI: http://localhost:6060/swagger-ui
4. See **internal/debug/debug.go** for more details

## ToDo List

1. Add tests (sqlmock)
2. Use https://github.com/Masterminds/squirrel library to generate DB query, at the moment query generation looks not so good
3. Add additional package to make the db connection, add db connection healthcheck (https://github.com/InVisionApp/go-health)
4. Add cache layer for the data