FROM golang:1.21 AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY templates/ ./templates
RUN CGO_ENABLED=0 GOOS=linux go build -o /dist/app

FROM build-stage AS run-stage
WORKDIR /dist
COPY --from=build-stage /dist/app ./
COPY --from=build-stage /app/templates/ ./templates/
EXPOSE 3000
CMD ["/dist/app"]
