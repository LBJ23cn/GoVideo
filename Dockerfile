# GoVideo 后端镜像：一个镜像同时包含 API 和 worker 两个入口，由 docker-compose.yml
# 通过 command 区分：
#   API    -> /app/api
#   Worker -> /app/worker
#
# 构建:
#   docker build -t govideo-app .

FROM golang:1.25-alpine AS build
WORKDIR /src

# 国内构建默认走 goproxy.cn；如需官方代理可传 --build-arg GOPROXY=https://proxy.golang.org,direct
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

# 先复制 go.mod/go.sum，依赖不变时可复用缓存层
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM alpine:3.21

# 默认使用阿里云 Alpine 源，避免国内拉取 dl-cdn.alpinelinux.org 过慢。
# 如需官方源可传 --build-arg APK_MIRROR=dl-cdn.alpinelinux.org
ARG APK_MIRROR=mirrors.aliyun.com
# ca-certificates: 访问阿里云 OSS（HTTPS）
# tzdata:          容器内时间按 TZ 设置
RUN sed -i "s|dl-cdn.alpinelinux.org|${APK_MIRROR}|g" /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

WORKDIR /app

COPY --from=build /out/api /out/worker /app/

# 视频/封面上传前会临时落盘到 .run/uploads，提前建好并交给非 root 用户
RUN mkdir -p /app/.run/uploads && chown -R app:app /app/.run

ENV TZ=Asia/Shanghai \
    CONFIG_PATH=configs/config-docker.yaml

USER app

EXPOSE 8080 6060 6061

CMD ["/app/api"]
