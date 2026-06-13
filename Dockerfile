# 多阶段构建：编译 Go 二进制并打包为最小运行镜像
FROM golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /wiki .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget \
	&& adduser -D -h /app appuser

WORKDIR /app

COPY --from=builder /wiki .
COPY manifest/config/ ./manifest/config/

RUN mkdir -p log && chown -R appuser:appuser /app

USER appuser

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
	CMD wget -qO- http://127.0.0.1:8081/health || exit 1

CMD ["./wiki", "prod"]
