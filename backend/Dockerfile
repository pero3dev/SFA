FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/sfa-api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=builder /out/sfa-api /app/sfa-api
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/sfa-api"]
