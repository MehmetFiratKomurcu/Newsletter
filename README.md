## Commands

### Run Postgres Docker Container
```json
docker run --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres
```

### Run OpenTelemetry
```json
docker run --rm --name jaeger -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 -p 16686:16686 -p 4318:4318 jaegertracing/all-in-one:1.56
```

 