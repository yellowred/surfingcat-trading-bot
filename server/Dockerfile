
FROM golang:1.9.2-alpine3.6

RUN apk add --no-cache curl && \
    curl -Lo /bin/rq  https://s3-eu-west-1.amazonaws.com/record-query/record-query/x86_64-unknown-linux-musl/rq && \
    chmod +x /bin/rq

RUN apk add --no-cache git openssl bzr \
    && go get -u github.com/golang/dep/cmd/dep

ENV GOPATH=/go
ENV SRC_DIR=${GOPATH}/src/github.com/yellowred/surfingcat-trading-bot/server/
RUN mkdir -p $SRC_DIR
WORKDIR $SRC_DIR

#ADD vendor /go/src/
#ADD . ${GOPATH}/src/github.com/yellowred/surfingcat-trading-bot/
ADD Gopkg.toml Gopkg.lock $SRC_DIR
# RUN cd /go/src && \
#  cat ${SRC_DIR}Gopkg.lock | /bin/rq -tJ 'map "projects" | spread | map "name"' | cat | tr -d '"' | xargs -I % go install %/...



RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep

RUN dep ensure -vendor-only

COPY . $SRC_DIR

#disable crosscompiling 
ENV CGO_ENABLED=0

#compile linux only
ENV GOOS=linux

RUN go install .
RUN go build -o server .

ENV API_PORT=3026
EXPOSE 3026

ENTRYPOINT ./server -kafka-consumer-group group.internal -kafka-host ${KAFKA_HOST}:9092 -kafka-zookeeper-host ${ZOOKEEPER_HOST}:2181 -mongo-host=${MONGO_HOST}:27017