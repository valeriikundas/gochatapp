FROM golang:1.21 AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
# RUN go mod download
COPY *.go dev_config.yaml templates/ ./
CMD ["go", "run", "."]
# RUN CGO_ENABLED=0 GOOS=linux go build -o /api

# FROM build-stage AS run-stage
# WORKDIR /
# COPY --from=build-stage /api /app/dev_config.yaml /api
# EXPOSE 3000
# CMD ["/api"]
