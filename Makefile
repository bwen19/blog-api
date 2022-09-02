DB_URL=postgresql://root:secret@localhost:5432/blog?sslmode=disable

network:
	docker network create blog-network

postgres:
	docker run --name postgres --network blog-network -p 5432:5432 -e POSTGRES_USER=root POSTGRES_PASSWORD=secret -d postgres:alpine

createdb:
	docker exec -it postgres createdb --username=root --owner=root blog

dropdb:
	docker exec -it postgres dropdb blog

db_schema:
	dbml2sql --postgres -o db/schema/schema.sql db/schema/blog.dbml

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

sqlc:
	cmd /C del db\sqlc\\*.sql.go
	docker run --rm -v D:\project\webapp\blog\server:/src -w /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

proto:
	cmd /C del pb\\*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb \
	--grpc-gateway_opt=paths=source_relative,allow_delete_body=true  \
	--openapiv2_out=swagger \
	--openapiv2_opt=allow_merge=true,merge_file_name=blog,allow_delete_body=true \
	proto/*.proto
	cmd /C del statik\\*.go
	statik -src=swagger

evans:
	evans --host localhost --port 9090 -r repl

.PHONY: network postgres createdb dropdb db_schema migrateup migratedown sqlc test server proto evans
