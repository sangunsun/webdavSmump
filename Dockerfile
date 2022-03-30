# build stage
FROM golang:1.17 AS builder

WORKDIR /app
COPY *.go ./
# RUN go mod download
RUN go build -o /webdavSmump
# 镜像是基于alpine
FROM alpine
# LABLE 给镜像添加元数据
# MAINTAINER 维护者信息
LABEL maintainer="sagit"
RUN mkdir /config
# ENV 指定环境变量
# 设置固定的项目路径
ENV WORKDIR /
# 添加执行文件
COPY --from=builder /app/webdavSmump /webdav
COPY entrypoint.sh /entrypoint.sh
# RUN 指令将在当前镜像基础上执行指定命令
# 添加应用可执行文件，并设置执行权限
RUN chmod 777 webdav
RUN chmod 777 entrypoint.sh
# EXPOSE docker容器暴露的端口
EXPOSE 8899
# 指定工作目录
WORKDIR /
VOLUME /config/config.json
# CMD 指定启动容器时执行的命令
ENTRYPOINT ["./entrypoint.sh"]