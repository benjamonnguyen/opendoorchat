# syntax=docker.io/docker/dockerfile:1.7-labs

FROM golang:1.22.1 AS build
WORKDIR /app
COPY --exclude=./cmd/frontend --exclude=./cmd/static_generator --exclude=./frontend . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /opendoor-chat-backend ./cmd/backend
RUN go test -v -skip IT ./...

FROM gcr.io/distroless/base-debian11 AS build-release
WORKDIR /
COPY --from=build /opendoor-chat-backend /opendoor-chat-backend
EXPOSE 8080

ENTRYPOINT ["/opendoor-chat-backend"]