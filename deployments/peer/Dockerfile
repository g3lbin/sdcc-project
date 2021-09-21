# syntax=docker/dockerfile:1
FROM golang:1.16 AS builder
ENV GO111MODULE=on
WORKDIR /go/src/github.com/sdcc-project
COPY . .
RUN go mod download && go build -o app internal/peer/peer.go

FROM golang:1.16-alpine
RUN apk update && apk add bash
WORKDIR /root/
COPY --from=builder /go/src/github.com/sdcc-project/app /go/src/github.com/sdcc-project/internal/peer/wait-for-it.sh ./
ENTRYPOINT [ "./wait-for-it.sh", "registration:${REGISTRY_PORT}", "-q" ]
CMD [  "--", "/bin/sh", "-c", "app" ]