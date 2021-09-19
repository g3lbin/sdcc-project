FROM golang:1.16-alpine

WORKDIR /go/src/app

COPY go.mod peer/wait-for-it.sh peer/peer.go pkg/lib/lib.go ./

RUN mkdir peer lib && mv peer.go peer && mv lib.go lib \
    && go mod download \
    && go build lib/lib.go \
    && go install ./peer \
    && apk update && apk add bash

ENTRYPOINT [ "./wait-for-it.sh", "registry:${REGISTRY_PORT}", "-q" ]
CMD [  "--", "sh", "-c", "/go/bin/peer" ]