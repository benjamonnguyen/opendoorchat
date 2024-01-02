FROM golang:1.21.4 AS build
WORKDIR /app
COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /opendoor-chat-services ./cmd/backend
RUN go test -v -skip IT ./...

FROM gcr.io/distroless/base-debian11 AS build-release
WORKDIR /
COPY --from=build /opendoor-chat-services /opendoor-chat-services
# TODO provide via volume
COPY ./config.yaml ./
EXPOSE 8080

ENTRYPOINT ["/opendoor-chat-services"]