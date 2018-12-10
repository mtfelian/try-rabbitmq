FROM golang

ENV GOBIN /go/bin

RUN mkdir /app
RUN mkdir -p /go/src/github.com/mtfelian
WORKDIR /go/src/github.com/mtfelian
RUN git clone https://github.com/mtfelian/try-rabbitmq
WORKDIR /go/src/github.com/mtfelian/try-rabbitmq

# Build my app
RUN go build -o /app/server ./server
WORKDIR /app
ENTRYPOINT ["/app/server"]
