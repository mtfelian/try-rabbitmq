FROM golang

ENV GOBIN /go/bin

RUN mkdir /app && \
  mkdir -p /go/src/github.com/mtfelian && \
  cd /go/src/github.com/mtfelian && \
  git clone https://github.com/mtfelian/try-rabbitmq && \
  cd /go/src/github.com/mtfelian/try-rabbitmq && \
  go build -o /app/server ./server

WORKDIR /app
ENTRYPOINT ["/app/server"]
