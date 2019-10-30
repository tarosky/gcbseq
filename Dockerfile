FROM golang:1.12 as builder

ARG GO_BASEDIR

ENV GO111MODULE on

COPY ["go.mod", "go.sum", "gcbseq.go", "/go/src/github.com/tarosky/gcbseq/"]
RUN cd /go/src/github.com/tarosky/gcbseq && go mod download
RUN cd /go/src/github.com/tarosky/gcbseq && go build -o /gcbseq gcbseq.go

FROM debian:stretch

COPY --from=builder ["/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs/ca-certificates.crt"]
COPY --from=builder ["/gcbseq", "/gcbseq"]

ENTRYPOINT ["/gcbseq"]
