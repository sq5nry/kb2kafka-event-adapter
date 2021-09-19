FROM golang:1.17-alpine as builder
WORKDIR /app/kafka-health-check-master

#TODO place sources into ci/cd nexus or ensure building phase has internet access
#the bundled Makefile is unused in favour of go modules
RUN wget https://github.com/andreas-schroeder/kafka-health-check/archive/refs/heads/master.zip && \
  unzip master.zip -d /app && \
  rm master.zip && \
  apk add build-base && \
  go mod init && \
  go mod vendor && \
  go build

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app .
CMD ["/app/kafka-health-check-master/kafka-health-check", \
    "-broker-id=1001", \
    "-broker-host=kafka", \
    "-broker-port=9092", \
    "-check-interval=10s", \
    "-server-port=8000", \
    "-zookeeper=zookeeper:2181"]