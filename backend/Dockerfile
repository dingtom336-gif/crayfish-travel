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

ENTRYPOINT ["/crayfish-server"]
