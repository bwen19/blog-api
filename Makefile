DB_URL=postgresql://root:secret@localhost:5432/blog?sslmode=disable
DOCKER_DB_URL=postgresql://root:secret@postgres:5432/blog?sslmode=disable
NETWORK=blog-network

network:
	docker network create "$(NETWORK)"

blog:
	docker run --name blog --network "$(NETWORK)" -p 8080:8080 -p 9090:9090 -e DB_SOURCE="${DOCKER_DB_URL}" blog:latest

redis:
	docker run --name redis --network "$(NETWORK)" -p 6379:6379 -d redis:alpine3.16

postgres:
	docker run --name postgres --network "$(NETWORK)" -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:14.5-alpine3.16

createdb:
	docker exec -it postgres createdb --username=root --owner=root blog

dropdb:
	docker exec -it postgres dropdb blog

schema:
	dbml2sql --postgres -o psql/schema/schema.sql psql/schema/blog.dbml

migrateup:
	migrate -path psql/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path psql/migration -database "$(DB_URL)" -verbose down

sqlc:
	rm -f psql/db/*.sql.go
	sqlc generate

proto:
	rm -f grpc/pb/*.pb.go
	protoc --proto_path=grpc/proto --go_out=grpc/pb --go_opt=paths=source_relative \
		--go-grpc_out=grpc/pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=grpc/pb \
		--grpc-gateway_opt=paths=source_relative,allow_delete_body=true \
		--openapiv2_out=grpc/swagger \
		--openapiv2_opt=allow_merge=true,merge_file_name=blog,allow_delete_body=true \
		grpc/proto/*.proto
	rm -f grpc/statik/statik.go
	statik -src=grpc/swagger -dest=grpc

test:
	go test -v -cover ./...

server:
	go run main.go

evans:
	evans --host localhost --port 9090 -r repl

.PHONY: network blog redis postgres createdb dropdb schema migrateup migratedown sqlc test server proto evans
