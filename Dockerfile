FROM golang:latest
MAINTAINER Gianluca Borello <g.borello@gmail.com>

COPY healthchecker /go/src/github.com/gianlucaborello/healthchecker-kubernetes/healthchecker
RUN go get -d -v github.com/gianlucaborello/healthchecker-kubernetes/healthchecker
RUN go install -v github.com/gianlucaborello/healthchecker-kubernetes/healthchecker

ENTRYPOINT ["/go/bin/healthchecker"]
