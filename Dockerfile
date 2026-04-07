FROM golang:1.24.3-alpine3.21 AS builder

WORKDIR /app
COPY . /app

# 调试信息
RUN go version
RUN go env GOPROXY
RUN ls -la

# 先下载依赖，再构建
RUN go mod download
RUN CGO_ENABLED=0 go build -v -o glance . || (echo "BUILD_FAILED" && cat /dev/null && exit 1)

FROM alpine:3.21

WORKDIR /app
COPY --from=builder /app/glance .

EXPOSE 8080/tcp
ENTRYPOINT ["/app/glance", "--config", "/app/config/glance.yml"]
