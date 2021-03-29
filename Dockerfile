FROM golang:1.16-alpine3.12

ENV GOPROXY="http://proxy.golang.org"
ENV GO111MODULE="on"

ENV PORT=8080
EXPOSE 8080

ENV RAFT_PORT=8081
EXPOSE 8081

WORKDIR /key-value-server

RUN apk add --no-cache git
COPY . .

RUN go build -v -o /go/bin/server

CMD ["/go/bin/server"]