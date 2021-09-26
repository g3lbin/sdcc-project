FROM golang:1.16-alpine

WORKDIR /go/src/github.com/sdcc-project

COPY go.mod go.sum internal/pkg/utils/wait-for-it.sh internal/sequencer/sequencer.go internal/pkg/rpcregistration/rpcregistration.go internal/pkg/rpcsequencer/rpcsequencer.go internal/pkg/utils/utils.go ./

RUN mkdir internal internal/sequencer internal/pkg internal/pkg/rpcregistration internal/pkg/rpcsequencer internal/pkg/utils \
    && mv sequencer.go internal/sequencer && mv rpcregistration.go internal/pkg/rpcregistration \
    && mv rpcsequencer.go internal/pkg/rpcsequencer && mv utils.go internal/pkg/utils \
    && go mod tidy && go mod download && go build internal/pkg/utils/utils.go \
    && go build internal/pkg/rpcregistration/rpcregistration.go \
    && go build internal/pkg/rpcsequencer/rpcsequencer.go && go install ./internal/sequencer \
    && apk update && apk add bash

ENTRYPOINT [ "./wait-for-it.sh", "registration:${REGISTRATION_PORT}", "-q" ]
CMD [  "--", "/go/bin/sequencer" ]