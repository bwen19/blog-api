postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root POSTGRES_PASSWORD=secret -d postgres:alpine

createdb:
	docker exec -it postgres createdb --username=root --owner=root blog

dropdb:
	docker exec -it postgres dropdb blog

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/blog?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/blog?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test
