FROM golang:1.25.1-alpine AS api_gateway_dev
RUN go install github.com/air-verse/air@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest
WORKDIR /app
COPY . .
RUN go mod download
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]


FROM api_gateway_dev AS build_stage
COPY . .
RUN swag init -g cmd/api-gateway/main.go -o docs
RUN CGO_ENABLED=0 GOOS=linux go build \
  -tags stage \
  -ldflags="-s -w" \
  -o /app/api-gateway \
  ./cmd/api-gateway

FROM api_gateway_dev AS build_prod
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
  -tags prod \
  -ldflags="-s -w" \
  -o /app/api-gateway \
  ./cmd/api-gateway


FROM alpine:latest AS api_gateway_stage
WORKDIR /app
COPY --from=build_stage /app/api-gateway .
COPY --from=build_stage /app/docs ./docs
RUN addgroup -g 1000 appuser && \
  adduser -D -u 1000 -G appuser appuser
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
CMD ["./api-gateway"]


FROM alpine:latest AS api_gateway_prod
WORKDIR /app
COPY --from=build_prod /app/api-gateway .
RUN addgroup -g 1000 appuser && \
  adduser -D -u 1000 -G appuser appuser
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
CMD ["./api-gateway"]