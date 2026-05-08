FROM golang:1.24-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/novel-reader-backend ./cmd/server

FROM alpine:3.22

WORKDIR /app
RUN apk add --no-cache ca-certificates && adduser -D -H appuser
COPY --from=builder /out/novel-reader-backend /app/novel-reader-backend
RUN mkdir -p /app/data/uploads && chown -R appuser:appuser /app
USER appuser
EXPOSE 8000
ENV HTTP_ADDR=:8000 UPLOAD_DIR=/app/data/uploads
ENTRYPOINT ["/app/novel-reader-backend"]
