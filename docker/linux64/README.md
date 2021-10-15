# 编译

```
docker build -t webdavsmump .
```

# 运行

```
docker run -v /root/github/webdavSmump/config.json:/config/config.json  -v /mnt/important/appbackup/:/mnt/important/appbackup/ -p 8899:8899 --name webdavtest webdavsmump:latest
```

`-v /root/github/webdavSmump/config.json:/config/config.json`   挂载/config/config.json配置文件，**注意：不要修改配置文件的端口**
`-v /mnt/important/appbackup/:/mnt/important/appbackup/`        挂载共享文件夹，**注意：容器内路径可以和主机一样，配置文件内的路径写容器内的**
`-p 8899:8899`                                                  端口映射**容器端口8899**

