FROM golang:1.21
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY templates/ ./templates
COPY docker_config.yaml ./
EXPOSE 3000
ENTRYPOINT ["go", "run", ".", "--config", "docker"]
