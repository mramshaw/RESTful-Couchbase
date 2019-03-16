FROM golang:1.12

RUN go get golang.org/x/lint/golint

RUN go get github.com/gorilla/mux
RUN go get gopkg.in/couchbase/gocb.v1

EXPOSE 8080
