FROM golang:1.16-alpine AS builder
WORKDIR /app
COPY . .
RUN rm -rf ./cmd ./test && go mod tidy && go mod download  && go build -o ./registration ./internal/registration

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/registration ./
ENTRYPOINT [ "/bin/sh", "-c", "./registration ${LISTENING_PORT}" ]