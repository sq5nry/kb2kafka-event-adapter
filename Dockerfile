FROM golang:1.17-alpine as builder
WORKDIR /app
COPY . .
RUN apk add build-base && \
    go build -tags musl -o adapter ./cmd/...

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/adapter .
CMD ["/app/adapter", "-alsologtostderr=true"]
