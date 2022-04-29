gen:
	 protoc --proto_path=proto --go_out=plugins=grpc:pb proto/*.proto
clean:
	rm pb/*.go
server:
	go run cmd/server/main.go -port 8080
client:
	go run cmd/client/main.go -addr="127.0.0.1:8080"
test:
	go test -cover ./...
