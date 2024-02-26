FROM golang:1.18 AS build
WORKDIR /src/mls
ENV GO111MODULE=on
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go install -a -tags netgo -ldflags=-w

FROM alpine:3.8
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN apk add --no-cache tzdata
ENV TZ UTC
COPY --from=build /go/bin/mlslisting /bin/mlslisting
COPY --from=build /src/mls/configs /configs
# for integration testing
COPY --from=build /src/mls/aws.local /var/secrets/aws.local


# Uncomment after chaging health check protocol implementation
#RUN GRPC_HEALTH_PROBE_VERSION=v0.3.1 && \
#    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
#   chmod +x /bin/grpc_health_probe
#HEALTHCHECK CMD ["/bin/grpc_health_probe", "-addr=:9080"]

ENTRYPOINT [ "/bin/mlslisting", "/configs" ]
## TODO: use env
EXPOSE 9080 9081 80
