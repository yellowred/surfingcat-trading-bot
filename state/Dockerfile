FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/sf-trading-bot-state

WORKDIR /go/src/sf-trading-bot-state

RUN go get ./...
RUN go install .

ENTRYPOINT /go/bin/sf-trading-bot-state -kafka-consumer-group group.internal -kafka-host 192.168.10.100:9092 -kafka-zookeeper-host 192.168.10.100:2181 -mongo-host 192.168.10.100:27017