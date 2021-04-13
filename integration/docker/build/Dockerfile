FROM golang:1.11

VOLUME ["/go"]

WORKDIR /go/src/theta/

ENV GOPATH=/go

ENV CGO_ENABLED=1 

ENV GO111MODULE=on

CMD ["/go/src/theta/integration/docker/build/start.sh"]



