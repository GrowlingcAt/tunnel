FROM quay.io/0voice/node:18.16.0 as stage0
RUN npm config set registry https://mirrors.huaweicloud.com/repository/npm/
COPY ./user-web /src/user-web
WORKDIR /src/user-web
RUN npm install
RUN npm run build

FROM quay.io/0voice/golang:1.20 as stage1
RUN go env -w GOPROXY=https://goproxy.cn,https://proxy.golang.com.cn,direct
ADD ./user /src/user
WORKDIR /src/user
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user .

FROM quay.io/0voice/alpine:3.18 as stage2
ADD ./curl-amd64 /usr/bin/curl
RUN chmod +x /usr/bin/curl
WORKDIR /app
ADD ./user/dev.config.yaml /app/config.yaml
COPY --from=stage0 /src/user-web/dist /app/www
COPY --from=stage1 /src/user/user /app
ENTRYPOINT ["./user"]
CMD ["--config=config.yaml"]


