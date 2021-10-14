# 镜像是基于alpine
FROM alpine:latest
# LABLE 给镜像添加元数据
# MAINTAINER 维护者信息
LABEL maintainer="sagit"
RUN mkdir /webdav
# ENV 指定环境变量
# 设置固定的项目路径
ENV WORKDIR /webdav
# ADD <src> <dest>  复制指定的 <src> 到容器中的 <dest>
# MyApp是Go代码生成的可执行文件
COPY webdavSmump_linux64 /webdav/main
# RUN 指令将在当前镜像基础上执行指定命令
# 添加应用可执行文件，并设置执行权限
RUN chmod +x /webdav/main
# 添加静态文件、配置文件、模板文件 (根据自己的项目实际情况配置)
# EXPOSE docker容器暴露的端口
EXPOSE 8899
# 指定工作目录
WORKDIR /webdav
#映射配置文件
VOLUME /webdav/config.json
# CMD 指定启动容器时执行的命令
ENTRYPOINT ["./main"]
