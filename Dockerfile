# 前端构建
FROM quay.io/0voice/node:18.16.0 AS stage0
RUN npm config set registry https://mirrors.huaweicloud.com/repository/npm/
COPY ./tunnel-web /src/tunnel-web
WORKDIR /src/tunnel-web
RUN npm install
RUN npm run build

# 后端构建
FROM golang:1.26 AS stage1
RUN go env -w GOPROXY=https://goproxy.cn,https://proxy.golang.com.cn,direct
ADD ./tunnel /src/tunnel
WORKDIR /src/tunnel
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o tunnel ./cmd/tunnel

FROM quay.io/0voice/alpine:3.18 AS stage2
MAINTAINER GrowlingcAt
WORKDIR /app/
ADD ./tunnel/dev.config.yaml /app/config.yaml
COPY --from=stage0 /src/tunnel-web/dist /app/www
COPY --from=stage1 /src/tunnel/tunnel /app
# 指定入口程序
ENTRYPOINT ["./tunnel"]
# 指定容器的启动命令或者入口程序的参数
CMD ["--config=config.yaml"]