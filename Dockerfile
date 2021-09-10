FROM golang:1.17-alpine as builder
WORKDIR /app
ADD . .
RUN apk add build-base && \
    go build -tags musl -o ratings ./cmd/..

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/ratings .
CMD ["/app/ratings"]
