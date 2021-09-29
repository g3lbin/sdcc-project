FROM golang:1.16-alpine AS builder
WORKDIR /app
COPY . .
RUN rm -rf ./cmd && go mod tidy && go mod download && go build -o sequencer ./internal/sequencer

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/sequencer /app/internal/pkg/utils/wait-for-it.sh ./
RUN  apk update && apk add bash
ENTRYPOINT [ "./wait-for-it.sh", "registration:${REGISTRATION_PORT}", "-q" ]
CMD [  "--", "./sequencer" ]