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
+ webdav的存取路径名和用户名一致，即webdav的访问路径是形如https://xx.xx.xx.xx:8899/username/ （用户名最后面的这个斜杠需要加上） ，配置文件不再体现

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

### webdav挂载
+ webdav挂载：davfs2,根据官方文档，只有root用户才能挂载，挂载后盘的所有者是root，造成普通用户没有写权限，需要chmod 777 dir修改目录权限后，普通用户才能正常访问(已验证)。
+ 已有解决办法如下：
```
sudo mount -t davfs -o uid=myUser -o gid=myUser https://example.org/remote.php/webdav/ /home/myUser/myDir/
```
> 1. mount -t davfs https://abc.com:8080/dav/ /mnt/webdav
> 2. mount -t davfs http(s)://addres:<port>/path /mount/point (官方文档提供)
> 3. 官方文档：https://wiki.archlinux.org/title/Davfs2
> 4. 配置文件:/etc/davfs2/davfs2.conf,

> 1. 遇到一个问题，使用davfs2挂载远程盘后，使用dd测试写盘速度，1G的文件瞬间写完成，但是到服务器上看，文件是0字节，umount盘后，提示在写缓存，以为稳了，umont成功后，到服务器看文件还是0字节。然后再挂载盘，想通过客户朵看远程文件情况，结果就一直卡，卡好后到服务器一看，文件已经是1G了。如果davfs2是这样的处理逻辑，那是不是很容易造成文件丢失？？？
> 2. 使用RaiDrive挂载盘，在拷贝大文件的时候，服务端的数据是实时增长的。
> 2. 这个结果说明davfs2并不是一个靠谱的挂载软件，RaiDrive是一个靠谱的挂载软件；同时说明服务端的webdav服务程序没问题，上述问题就是davfs2造成的。

> 挂载webdav的另一种方法使用wdfs - webdav filesystem for fuse，见：https://serverfault.com/questions/391717/mounting-webdav-as-user-no-sudo

+ 开机自动挂载(该方法还未验证)
> 1. 启用 davfs2 用户锁:将配置文件中use_locks前面的#去掉，并将1改为0，保存退出可重启自动挂载。
> 2. 不用每次挂载输用户名密码(挂载自动提供用户名密码)：修改 /etc/davfs2/secrets,末尾增加形如下面一行 https://abc.com:8080/dav/ userxxx passwordxxxx(该方法可用性已验证)
> 3. 开机自动挂载：编辑/etc/rc.local，在文件末尾加一行: mount -t davfs https://abc.com:8080/dav/ /mnt/webdav
+ webdav挂载：fusedav
> fusedav https://webdav链接 /home/test/ -u (你的账户名) -p (你的密码)

+ 又一个挂载软件Rclone，一个使用指南：https://www.moewah.com/archives/876.html

#### 手机版wps挂载webdav说明
+ https://blog.jianguoyun.com/?p=2576

#### windows挂载webdav
+ windows下默认只支持https方式的webdav挂载，并且要求是可信任的证书，在局域网中这样使用就很麻烦，使用http方式更合适。所以windows下挂载webdav共三步

1. 开启服务:windows下webdav的客户端程序是webclient服务，在电脑服务中启动次服务，并设置为自动启动运行。

2. 修改注册表
修改注册表使得WIN同时支持http和https：

    > 定位到
HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\WebClient\Parameters
把BasicAuthLevel 值改成2，即同时支持http和https，默认只支持https，
然后重启服务：

+ 映射磁盘
    >使用windows的添加一个网络位置功能
填写正确的连接加端口号，共享目录名可以不填写，也可以填写，如果填写千万不要填写错了。

+ windows自带客户端无法下载webdav服务器上大于50M文件的问题
> https://support.microsoft.com/zh-cn/topic/%E4%BB%8E-web-%E6%96%87%E4%BB%B6%E5%A4%B9%E4%B8%8B%E8%BD%BD%E5%A4%A7%E4%BA%8E-50000000-%E5%AD%97%E8%8A%82%E7%9A%84%E6%96%87%E4%BB%B6%E6%97%B6%E5%87%BA%E7%8E%B0%E6%96%87%E4%BB%B6%E5%A4%B9%E5%A4%8D%E5%88%B6%E9%94%99%E8%AF%AF%E6%B6%88%E6%81%AF-815e2949-0f56-ec25-db7d-b6d860a31f77
+ windows客户端最大只能下载4G以下文件的问题？？？
 
#### 使用RaiDrive挂载
    + 使用RaiDrive挂载可在windows下挂载大多数协议的网盘，包括webdav、ftp等
 
### TODO
1. 拟增加浏览器查看markdown文件时，直接把markdown文档.md文档渲染显示功能。
