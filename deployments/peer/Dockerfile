FROM golang:1.16-alpine

WORKDIR /go/src/github.com/sdcc-project

COPY go.mod go.sum internal/pkg/utils/wait-for-it.sh internal/peer/peer.go internal/pkg/rpcregistration/rpcregistration.go internal/pkg/rpcsequencer/rpcsequencer.go internal/pkg/utils/utils.go ./

RUN mkdir internal internal/peer internal/pkg internal/pkg/rpcregistration internal/pkg/rpcsequencer internal/pkg/utils \
    && mv peer.go internal/peer && mv rpcregistration.go internal/pkg/rpcregistration \
    && mv rpcsequencer.go internal/pkg/rpcsequencer && mv utils.go internal/pkg/utils \
    && go mod tidy && go mod download  && go build internal/pkg/utils/utils.go \
    && go build internal/pkg/rpcregistration/rpcregistration.go \
    && go build internal/pkg/rpcsequencer/rpcsequencer.go && go install ./internal/peer \
    && apk update && apk add bash

ENTRYPOINT [ "./wait-for-it.sh", "registration:${REGISTRATION_PORT}", "-q" ]
CMD [  "--", "/go/bin/peer" ]