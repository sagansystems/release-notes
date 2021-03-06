FROM golang:1.10-alpine3.8 AS builder

ARG gopath="/go"
ENV GOPATH=${gopath}
ENV PROJECT_DIR=$GOPATH/src/github.com/sagansystems/release-notes
WORKDIR $PROJECT_DIR

COPY . .

RUN go test && go build -o release-notes

FROM alpine:3.8

ADD entrypoint.sh /

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/sagansystems/release-notes/release-notes /release-notes

ENTRYPOINT ["/entrypoint.sh"]