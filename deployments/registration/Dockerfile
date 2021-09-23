FROM golang:1.16-alpine

WORKDIR /go/src/github.com/sdcc-project

COPY go.mod go.sum internal/registration/registration.go internal/pkg/lib/lib.go ./

RUN mkdir internal internal/registration internal/pkg internal/pkg/lib && mv registration.go internal/registration && mv lib.go internal/pkg/lib \
    && go mod tidy && go mod download && go build internal/pkg/lib/lib.go && go install ./internal/registration

ENTRYPOINT [ "/bin/sh", "-c", "/go/bin/registration ${LISTENING_PORT}" ]