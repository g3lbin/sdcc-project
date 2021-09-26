FROM golang:1.16-alpine

WORKDIR /go/src/github.com/sdcc-project

COPY go.mod go.sum internal/registration/registration.go internal/pkg/rpcregistration/rpcregistration.go internal/pkg/utils/utils.go ./

RUN mkdir internal internal/registration internal/pkg internal/pkg/rpcregistration internal/pkg/utils \
    && mv registration.go internal/registration && mv rpcregistration.go internal/pkg/rpcregistration \
    && mv utils.go internal/pkg/utils \
    && go mod tidy && go mod download  && go build internal/pkg/utils/utils.go \
    && go build internal/pkg/rpcregistration/rpcregistration.go && go install ./internal/registration

ENTRYPOINT [ "/bin/sh", "-c", "/go/bin/registration ${LISTENING_PORT}" ]