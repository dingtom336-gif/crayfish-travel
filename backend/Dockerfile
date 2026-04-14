# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /crayfish-server ./cmd/server

# Stage 2: Run
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /crayfish-server /crayfish-server

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ready || exit 1

ENTRYPOINT ["/crayfish-server"]
