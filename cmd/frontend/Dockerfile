FROM golang:1.21.4 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /opendoor-chat-frontend

# FROM build as test
# RUN go test -v ./...

# FROM gcr.io/distroless/base-debian11 AS build-release
# WORKDIR /
# COPY --from=build /opendoor-chat-frontend /opendoor-chat-frontend
EXPOSE 3000

ENTRYPOINT ["/opendoor-chat-frontend"]