FROM golang:alpine

RUN apk add --no-cache git mercurial \
    && go get github.com/lib/pq \
    && go get github.com/jmoiron/sqlx \
    && apk del git mercurial

COPY ./src /go/src/app
WORKDIR /go/src/app

ENTRYPOINT ["go", "run", "main.go"]