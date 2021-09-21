# syntax=docker/dockerfile:1
FROM golang:1.16 AS builder
ENV GO111MODULE=on
WORKDIR /go/src/github.com/sdcc-project
COPY . .
RUN go mod download && go build -o app internal/registration/registration.go

FROM golang:1.16-alpine
WORKDIR /root/
COPY --from=builder /go/src/github.com/sdcc-project/app ./
ENTRYPOINT [ "/bin/sh", "-c", "app ${LISTENING_PORT}" ]