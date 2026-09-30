FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go openapi.yaml docs.html ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /matrix-api .

FROM scratch
COPY --from=build /matrix-api /matrix-api
USER 10001
EXPOSE 8080
ENTRYPOINT ["/matrix-api"]
