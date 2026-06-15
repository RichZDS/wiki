# ---- 前端构建阶段 ----
FROM node:20-alpine AS ui-builder

WORKDIR /ui
COPY ui/package.json ui/package-lock.json ./
RUN npm ci

COPY ui/ ./
RUN npm run build

# ---- Go 构建阶段 ----
FROM golang:1.25-alpine AS go-builder

RUN apk --no-cache add ca-certificates

WORKDIR /src

# 先复制依赖文件，利用 Docker 层缓存加速
COPY svr/go.mod svr/go.sum ./
RUN go mod download

# 复制源码并编译
COPY svr/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /wiki .

# ---- 运行阶段 ----
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 复制 Go 编译产物与配置文件
COPY --from=go-builder /wiki .
COPY svr/manifest/ ./manifest/

# 复制前端构建产物
COPY --from=ui-builder /ui/dist/ ./public/

EXPOSE 8081

CMD ["./wiki", "prod"]
