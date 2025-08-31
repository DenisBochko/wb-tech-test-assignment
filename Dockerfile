# Stage 1: Builder
FROM golang:1.24.4 AS builder

WORKDIR /app

COPY . .

RUN GOOS=linux go build -o /app/bin/wb-tech-test-assignment ./cmd/wb_tech_test_assignment

# Stage 2: Run
FROM ubuntu:latest

WORKDIR /app

COPY --from=builder /app/bin/wb-tech-test-assignment /app/bin/wb-tech-test-assignment
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/templates /app/templates

EXPOSE 8080

CMD ["/app/bin/wb-tech-test-assignment", "--config=/app/config/config.docker.yml"]
