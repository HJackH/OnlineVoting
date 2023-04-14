pb:
	protoc 	--go_out=./voting/ --go_opt=paths=source_relative \
			--go-grpc_out=./voting/ --go-grpc_opt=paths=source_relative \
			voting.proto

c:
	go run client/main.go

s:
	go run server/main.go

