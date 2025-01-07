# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o drive-checker .

# Runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/drive-checker .
COPY --from=builder /app/.env .

ENV PORT=3000
EXPOSE 3000

CMD ["./drive-checker"]
