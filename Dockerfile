FROM golang:1.17-alpine as builder
WORKDIR /app
ADD . .
RUN go build -o recipes ./cmd/...

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/recipes .
CMD ["/app/recipes"]
