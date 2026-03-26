# 阶段1：构建前端
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# 阶段2：构建后端
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY master/ .
ENV CGO_ENABLED=0
ENV GOTOOLCHAIN=auto
RUN go build -o jumpfrp-master ./cmd/server

# 阶段3：运行
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=backend-builder /app/jumpfrp-master .
COPY --from=frontend-builder /app/dist ./web
COPY scripts/ ./scripts/
RUN mkdir -p /data
EXPOSE 8080
CMD ["./jumpfrp-master"]
