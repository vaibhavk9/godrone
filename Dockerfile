FROM golang:1.7.1-alpine



ENV NAME=rest-registration

ENV DIR=/go/src/github.com/zang-cloud/$NAME


# Copy the local package files to the container's workspace.


ENV ADDR ":8889"

ENV GRPC_SERVICE_AUTH_ENDPOINT "micro-registration-auth:8888"


#TO USE ACCOUNT MOCK
#ENV ACCOUNTS_MOCK "true"

EXPOSE $ADDR


CMD $NAME


ADD . $DIR

WORKDIR $DIR

RUN go build && go install
