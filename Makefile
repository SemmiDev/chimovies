run:
	go run cmd/api/*.go
up:
	docker-compose up -d
destroy:
	docker rm -f postgres
