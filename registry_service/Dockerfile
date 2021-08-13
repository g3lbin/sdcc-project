FROM golang:1.16-alpine

WORKDIR /go/src/app

COPY registry/registry.go lib/lib.go ./

RUN mkdir registry lib && mv registry.go registry && mv lib.go lib \
    && go mod init && go mod tidy && go mod download \
    && go build lib/lib.go \
    && go install ./registry

ENTRYPOINT [ "sh", "-c", "/go/bin/registry" ]