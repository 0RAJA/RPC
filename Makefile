gen:
	protoc --proto_path=proto --go_out=plugins=grpc:pb proto/*.proto --grpc-gateway_out=:pb --swagger_out=:swagger
clean:
	rm pb/*.go
rest:
	go run cmd/server/main.go -port 8081 -type rest -endpoint 0.0.0.0:8080
server:
	go run cmd/server/main.go -port 8080
server1:
	go run cmd/server/main.go -port 50051
server2:
	go run cmd/server/main.go -port 50052
server1_tls:
	go run cmd/server/main.go -port 50051 -tls
server2_tls:
	go run cmd/server/main.go -port 50052 -tls
client:
	go run cmd/client/main.go -addr="127.0.0.1:8080"
client_tls:
	go run cmd/client/main.go -addr="127.0.0.1:8080" -tls
test:
	go test -cover ./...
enans:
	evans -r repl -p 8080 # evans cli
cert:
	cd cert; ./gen.sh; cd ..
.PHONY: clean server client test enans cert
