# syntax=docker.io/docker/dockerfile:1.7-labs

FROM golang:1.21.4 AS build
WORKDIR /app
COPY --exclude=./cmd/backend --exclude=./backend --exclude=./cmd/static_generator . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /opendoor-chat-frontend ./cmd/frontend
RUN go test -v -skip IT ./...

FROM gcr.io/distroless/base-debian11 AS build-release
WORKDIR /
COPY --from=build /opendoor-chat-frontend /opendoor-chat-frontend
EXPOSE 3000

ENTRYPOINT ["/opendoor-chat-frontend"]