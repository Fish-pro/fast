protoc:
	 cd api/proto/v1 && protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        *.proto

codegen:
	hack/update-codegen.sh

crdgen:
	hack/update-crdgen.sh