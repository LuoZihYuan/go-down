FROM golang:1.25.1-alpine AS payment_service_dev
RUN go install github.com/air-verse/air@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest
WORKDIR /app
COPY . .
RUN go mod download
EXPOSE 8082
CMD ["air", "-c", ".air.toml"]


FROM payment_service_dev AS build_stage
COPY . .
RUN swag init -g cmd/payment-service/main.go -o docs
RUN CGO_ENABLED=0 GOOS=linux go build \
  -tags stage \
  -ldflags="-s -w" \
  -o /app/payment-service \
  ./cmd/payment-service

FROM payment_service_dev AS build_prod
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
  -tags prod \
  -ldflags="-s -w" \
  -o /app/payment-service \
  ./cmd/payment-service


FROM alpine:latest AS payment_service_stage
WORKDIR /app
COPY --from=build_stage /app/payment-service .
COPY --from=build_stage /app/docs ./docs
RUN addgroup -g 1000 appuser && \
  adduser -D -u 1000 -G appuser appuser
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8082
CMD ["./payment-service"]


FROM alpine:latest AS payment_service_prod
WORKDIR /app
COPY --from=build_prod /app/payment-service .
RUN addgroup -g 1000 appuser && \
  adduser -D -u 1000 -G appuser appuser
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8082
CMD ["./payment-service"]