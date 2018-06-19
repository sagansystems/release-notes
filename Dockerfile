FROM golang:alpine AS builder

ARG gopath="/go"
ENV GOPATH=${gopath}
ENV PROJECT_DIR=$GOPATH/src/github.com/sagansystems/release-notes
WORKDIR $PROJECT_DIR

COPY . .

RUN go build -o release-notes

FROM alpine:3.7

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/sagansystems/release-notes/release-notes /release-notes

ENTRYPOINT ["/release-notes"]