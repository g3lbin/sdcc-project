FROM golang:1.16-alpine

WORKDIR /go/src/github.com/sdcc-project

COPY go.mod go.sum internal/peer/wait-for-it.sh internal/peer/peer.go internal/pkg/lib/lib.go ./

RUN mkdir internal internal/peer internal/pkg internal/pkg/lib && mv peer.go internal/peer && mv lib.go internal/pkg/lib \
    && go mod tidy && go mod download && go build internal/pkg/lib/lib.go && go install ./internal/peer \
    && apk update && apk add bash

ENTRYPOINT [ "./wait-for-it.sh", "registration:${REGISTRY_PORT}", "-q" ]
CMD [  "--", "/go/bin/peer" ]