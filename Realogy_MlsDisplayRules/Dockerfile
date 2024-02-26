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
COPY --from=build /go/bin/mls-display-rules /bin/mls-display-rules

EXPOSE 8083 9092 9981
CMD ["/bin/mls-display-rules"]

#FROM alpine:edge
#ADD mls-display-rules /
#
#RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
#EXPOSE 8083 9092 9981
#CMD ["/mls-display-rules"]