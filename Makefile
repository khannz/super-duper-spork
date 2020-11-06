build: gen-proto build-server

run-server: gen-proto
	go run server/server.go

build-server: gen-proto
	go build -o build/server server/server.go

gen-proto: clean
	@mkdir -p sdslogic
	@protoc \
	--proto_path=./proto/ \
	--go_out=sdslogic/ \
	--go-grpc_out=sdslogic/ \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative \
	sds.proto

clean:
	@rm -rf sdslogic/ build/