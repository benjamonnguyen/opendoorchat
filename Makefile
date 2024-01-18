unittest:
	go test -skip IT -cover -v ./...

# BACKEND

build-backend:
	docker build -t opendoor-chat-backend .

start-backend:
	go run ./cmd/backend

# FRONTEND 

dev-frontend:
	air -c ./.air.toml

tmpl-gen:
	templ generate
	go run ./cmd/static_generator

start-frontend: tmpl-gen
	go run ./cmd/frontend

#

start-tc:
	docker compose -f ./test-containers.yml up
