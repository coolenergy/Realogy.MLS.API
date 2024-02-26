#!/bin/sh
PROTO_ROOT_PATH=$1
API_VERSION=$2
GRPC_GATEWAY_VERSION=2.10.0


echo 'PROTO_ROOT_PATH : '$1
echo 'API_VERSION : '$2

export APP_ROOT_PATH=${PWD}
export GENERATED_PATH=internal/generated
export GO_STUB_PATH=$GENERATED_PATH/realogy.com/api/mls/

echo "# Installing tools and plugins"
GOOS=linux go get -u -d \
  github.com/golang/mock/mockgen@v1.5.0 \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
  google.golang.org/protobuf/cmd/protoc-gen-go \
  google.golang.org/grpc/cmd/protoc-gen-go-grpc \
  github.com/srikrsna/protoc-gen-gotag \
  github.com/grpc-ecosystem/grpc-gateway/v2@v${GRPC_GATEWAY_VERSION}

echo '\t1. Generating go server stub, gateway stub and swagger-openapi spec'
protoc -I/usr/local/include -I. \
          -I$GOPATH/pkg/mod \
          --go_out $GENERATED_PATH \
          --go-grpc_out $GENERATED_PATH \
          --grpc-gateway_out=logtostderr=true:$GENERATED_PATH \
          --openapiv2_out=json_names_for_fields=true,logtostderr=true:./api/swagger \
          --proto_path=$PROTO_ROOT_PATH realogy/api/mls/$API_VERSION/mls_listing.proto realogy/api/mls/$API_VERSION/tagger.proto

echo '\t2. Removing "xxx" fields and add appropriate struct tags'
cd $GENERATED_PATH
protoc -I/usr/local/include -I. \
            -I$GOPATH/pkg/mod \
            --gotag_out=xxx="graphql+\"-\" bson+\"-\" json+\"-\"":. \
            --proto_path=$PROTO_ROOT_PATH realogy/api/mls/$API_VERSION/mls_listing.proto

#echo '\t3. Generating mocks for test'
#cd $APP_ROOT_PATH
#mockgen -package=mock -destination=internal/services/mock/mock_services.go mlslisting/internal/generated/realogy.com/api/mls/v1 MlsListingServiceClient