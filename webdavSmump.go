package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
	"golang.org/x/net/webdav"
)

var servicePort int64
var prefixDir string
var configFileName = "config.json"
var jsonData = ""
var readFileTicker = time.NewTicker(10 * time.Second)
var readFileLock sync.Mutex

func httpHandler(w http.ResponseWriter, req *http.Request) {

	// 获取用户名/密码
	userName, password, ok := req.BasicAuth()
	//log.Println("userName:", userName, "password:", password)
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 验证用户名/密码

	//用锁对多协程读写配置文件资源进行同步----------
	select {
	case <-readFileTicker.C:
		readFileLock.Lock()
		jsonData, _ = getStringFromFile(configFileName)
		readFileLock.Unlock()
	default:
		readFileLock.Lock()
		readFileLock.Unlock()
	}

	user := gjson.Get(jsonData, "users.#(username="+userName+")#")
	//log.Println("user:", user)
	if !user.Exists() {
		http.Error(w, "WebDAV: need authorized!", http.StatusUnauthorized)
		//log.Println("no user")
		return
	}

	//判断用户口令是否正确，口令可以直接存储，也可以以md5码存储，这里程序进行自动判断。
	//防止把md5码当明码进行对比，不允许密码长度为32个字符
	if len(password) == 32 {
		http.Error(w, "WebDAV: need authorized!", http.StatusUnauthorized)
		return
	}
	userPath := gjson.Get(user.String(), "#(password="+password+").userpath")
	if !userPath.Exists() {
		passtemp := md5.Sum([]byte(password))
		//不能直接md5.Sum([]byte(password))[:] ,会报slice of unaddressable value错误
		userPath = gjson.Get(user.String(), "#(password="+hex.EncodeToString(passtemp[:])+").userpath")
		if !userPath.Exists() {
			http.Error(w, "WebDAV: need authorized!", http.StatusUnauthorized)
			//log.Println("wrong password")
			return

		}
	}

	//log.Println("userpath:", userPath)

	webdavDir := webdav.Dir(prefixDir + userPath.String())
	fs := &webdav.Handler{
		Prefix:     "/" + userName, //http传过来的目录名必须和用户名相同
		FileSystem: webdavDir,      //服务器上对应的目录路径是固定的目录前缀+用户名,目录前缀必须以符号/结束。
		LockSystem: webdav.NewMemLS(),
	}

	//log.Println("fs.Prefix:", fs.Prefix,"localpath:", webdavDir)
	if strings.HasPrefix(req.RequestURI, fs.Prefix) {
		fs.ServeHTTP(w, req)
		//log.Println("fs call")
		return
	}

	w.WriteHeader(404)
}

func webDavLoad() {
	/*
	   1. 读服务端口和目录前缀
	   2. 开启web服务
	*/
	jsonData, _ = getStringFromFile(configFileName)
	servicePort = gjson.Get(jsonData, "serviceport").Int()
	prefixDir = gjson.Get(jsonData, "prefixdir").String()

	s_mux := http.NewServeMux()
	s_mux.HandleFunc("/", httpHandler)

	dav_addr := fmt.Sprintf(":%v", servicePort)
	log.Println("webDav Server run ", dav_addr)

	//读ca证书的公钥、私钥，如读成功，启动https服务，如果读不成功启动http服务
	caKey := gjson.Get(jsonData, "cakey")
	caCrt := gjson.Get(jsonData, "cacrt")
	log.Println(caKey.String(), caCrt.String())
	if !caKey.Exists() || !caCrt.Exists() {
		log.Println("webDav server run http mode")
		//http.ListenAndServe是阻塞语句，不出错，后面的语句不会执行
		err := http.ListenAndServe(dav_addr, s_mux)
		if err != nil {
			log.Println("webDav server run error:", err)
		}
	} else {
		log.Println("webDav server run  https mode")
		//http.ListenAndServeTLS是阻塞语句，不出错，后面的语句不会执行
		err := http.ListenAndServeTLS(dav_addr, caCrt.String(), caKey.String(), s_mux)
		if err != nil {
			log.Println("webDav server run error:", err)
		}
	}

}

func main() {
	/*
	   1. 用一个go协程开启管理端口服务(具体服务程序放后期开发完善)
	   2. 开启webdav服务
	*/

	webDavLoad()

}

//从文件读入数据为字符串
func getStringFromFile(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		log.Println("打开文件失败:", err)
		return "", err
	}
	defer f.Close()
	fd, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("ioutil 读取文件失败:", err)
		return "", err
	}
	return string(fd), nil
}
