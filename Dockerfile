FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY master/ .
ENV CGO_ENABLED=0
ENV GOTOOLCHAIN=auto
RUN go build -o jumpfrp-master ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/jumpfrp-master .
COPY scripts/ ./scripts/
RUN mkdir -p /data /app/web
EXPOSE 8080
CMD ["./jumpfrp-master"]
