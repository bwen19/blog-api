DB_URL=postgresql://root:secret@localhost:5432/blog?sslmode=disable

network:
	docker network create blog-network

postgres:
	docker run --name postgres --network blog-network -p 5432:5432 -e POSTGRES_USER=root POSTGRES_PASSWORD=secret -d postgres:alpine

createdb:
	docker exec -it postgres createdb --username=root --owner=root blog

dropdb:
	docker exec -it postgres dropdb blog

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go blog/server/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock
