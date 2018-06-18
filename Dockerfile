FROM golang:alpine AS builder

ARG gopath="/go"

ENV GOPATH=${gopath}

COPY *.go $GOPATH/

RUN go build -o release-notes

FROM alpine:3.7

ADD entrypoint.sh /

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/release-notes /release-notes

ENTRYPOINT ["/entrypoint.sh"]