FROM golang:1.7.1-alpine

ENV NAME=sample

ENV DIR=/go/src/github.com/zang-cloud/$NAME

CMD $NAME

ADD . $DIR

WORKDIR $DIR

RUN go build && go install
