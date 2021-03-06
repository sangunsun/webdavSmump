# webdavSmump
a webdav server of mutil user mutil path

+ 因公司文件共享需求，寻找一个文件共享服务端软件，找了一圈发现现存的开源webdav服务软件极少实现了不同用户访问不同共享文件夹的。找现成的反倒不如自己写一个快，就写了本项目。原需求如下：

## 需求
总体描述：需要一个多用户文件共享方案，以满足企业单位员工在单位和互联网上共享文件的需要

### 需求点收集
1. 每一个用户都有自己的私人文件夹，部门共享文件夹，企业公共文件夹三个基本文件夹(也可使用挂载三个盘解决)
2. 可从单位内部和互联网访问文件夹
3. 方面的图形化用户及目录管理界面（还未实现，目前通过修改配置文件管理用户及用户访问的目录）
4. 可方便的在windows、linux下映射挂载为盘，移动终端上有app可以访问
5. 加密传输

## 解决方案

### 传输协议选择
原则:采用通用传输协议
#### 淘汰的协议
1. 因互联网传输安全性问题淘汰ftp协议
2. 因安全性问题淘汰smb协议
3. 因管理及挂载技术原因淘汰iscsi协议
4. 因管理及挂载技术原因淘汰NFS协议

#### 备选协议
1. sftp
2. webdav

### 服务软件选择

#### 第一选择：支持webdav多用户管理的服务软件
seafile 可道云 群晖drive、https://github.com/hacdias/webdav（二进制版本在树莓派多连接传输时会崩溃）
+ 淘汰原因：系统复杂，安装麻烦，有一定的学习曲线或不稳定

#### 第二选择：linux内置的sftp服务软件
+ 淘汰原因，用户管理和linux系统绑定，权限管理麻烦。

#### 第三选择：定制开发webdav服务软件
+ 选择原因：代码小，实现快
+ golang自带的webdav开发包帮助文档：https://pkg.go.dev/golang.org/x/net/webdav
+ webdav协议：http://www.webdav.org/specs/rfc2518.html
+ go语言提供的webdav支持:golang.org/x/net/webdav

## 最终选择自已定制开发webdav服务软件即本系统。
## 系统特点
1. 安装简单，如果以https方式运行就一个可执行文件、一个配置文件，一个公钥文件、一个私钥文件；如果是以http方式运行只需要前两个文件就可以正常运行。
2. 支持多用户登录，不同的用户访问不同的服务器文件夹
3. 加密通讯
4. 方便互联网和内部访问
5. 可运行在windows、linux、树莓派、macos等几乎所有操作系统下(只要golang支持的操作系统都可以运行）
6. 性能较好，文件传输输快，可同时传输多个文件。
7. 在webdav中用户名(username),访问路径(URL),服务器上的文件路径(userpath)是一个多对多对多的关系,实现起来比较复杂。本项目暂把这个关系简化成了一对一对一的关系。

## 安装指南
### 配置文件说明：
```
{
    "serviceport":8899,
    "cakey":"server.key",
    "cacrt":"server.crt",
    "prefixdir":"/srv/dev-disk-by-uuid-79746475560579FF/",
    "users":[
    {"username":"abc","password":"123","userpath":"minaaa"},
    {"username":"abc2","password":"202cb962ac59075b964b07152d234b70","userpath":"abc2"}
    ]
}
```
+ serviceport:服务端口
+ cakey:私钥证书存取路径
+ cacrt:公钥证书存取路径
+ prefixdir:服务器端共享文件夹的前缀，和userpath合起来组成共享给某个用户的文件夹在服务器上的绝对路径
+ users:用户集合，每一行对应一个用户描述，分别是用户名、用户口令(可以是明文也可以放置经md5哈稀后的口令)、服务器上分配给用户的共享文件夹名
+ webdav的存取路径名和用户名一致(即webdav的访问路径是形如https://xx.xx.xx.xx:8899/username) ，配置文件不再体现。

### 安装过程
1. 下载并编译主程序文件
2. 把配置文件config.json和主程序文件放入同一文件夹中
3. 在同一目录内放置供https通讯使用的公钥文件和私钥文件(没公钥文件和私钥文件也没关系，系统会转为http方式运行)
4. 按实际情况编辑好配置文件config.json并保存
5. 运行主程序文件
6. 用任一webdav客户端软件访问本服务程序(直接用浏览器访问会返回"Method Not Allowed",另外windows下的"添加一个网络位置"功能也不能正常访问)
+ **记得在用户访问前要把配置文件中userpath表示的目录创建好，否则用户访问的时候由于系统无法找到目录，会提示用户路径无法找到**

### 已知webdav缺点：
1. 客户端挂载问题：实测在linux下和在macos下，把webdav挂载成盘后，不管是命令行还是GUI拷贝文件都存在缓存问题，数据不能及时到达服务端。客户端提示文件拷贝完成，服务端看到文件大小还是0，即使umount掉盘数据还是不能及时到达服务器造成数据丢失。使用手机端的各种文件管理器、raidrive未发现丢失数据情况，同样一块远程共享，同时用webdav和smb共享，linux和macos下挂盘后拷贝数据，smb可把数据实时传给服务器，webdav就不行
2. 延迟问题：实测把同样的服务端目录通过smb和webdav共享出来，用同一个客户端的两种协议访问，在文件夹内文件较多(nnn以上)时，smb反应较快，而webdav则有一到几秒的延迟时间才能显示文件列表。
3. 在线播放问题(使用FE文件游览器)：和问题2一样的环境，在线播放nnnM的视频文件，smb基本能秒出，webdav则需要几秒的缓存时间。
