# Dockerfile for Swilly Webhook Server (File Processor)
FROM golang:latest

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

# Build the Go app
# RUN CGO_ENABLED=0 GOOS=linux go build -a -o go-api .
RUN --mount=type=cache,target=/root/.cache/go-build go build -o notifier ./alert_notifier/.

RUN cp notifier /root/

WORKDIR /root/

CMD ["./notifier"]
