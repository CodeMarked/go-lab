FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go_CRUD_api/go.mod go_CRUD_api/go.sum ./
RUN go mod download
COPY go_CRUD_api/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/go-crud-api .

FROM alpine:3.20
WORKDIR /app
RUN adduser -D -u 10001 appuser
COPY --from=build /out/go-crud-api /app/go-crud-api
USER appuser
EXPOSE 5000
ENTRYPOINT ["/app/go-crud-api"]
