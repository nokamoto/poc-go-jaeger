
proto:
	protoc --go_out=plugins=grpc:service -I protobuf protobuf/*.proto

fmt:
	gofmt -d .
	gofmt -w .

go: fmt proto
	dep ensure
	go test .

docker: go
	docker-compose down
	docker-compose build
