FROM golang:1.20.5 AS builder
ENV GOBIN /go/bin
RUN mkdir /app && \
  mkdir -p /go/src/github.com/mtfelian && \
  cd /go/src/github.com/mtfelian && \
  git clone https://github.com/mtfelian/try-rabbitmq && \
  cd /go/src/github.com/mtfelian/try-rabbitmq/server && \
  go build -o /app/server

FROM debian
RUN apt update && apt upgrade -y && mkdir /app
COPY --from=builder /app/server /app/server
WORKDIR /app
ENTRYPOINT ["/app/server"]
