FROM golang:1.16-alpine

RUN export GOPATH=$(go env GOPATH)
WORKDIR $GOPATH/src/app

COPY registry/registry.go ./registry/
COPY lib ./lib

RUN go mod init \
    && go mod tidy \
    && go mod download \
    && go build lib/lib.go \
    && go install ./registry

CMD [ "sh", "-c", "${GOPATH}/bin/registry" ]