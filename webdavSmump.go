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
	"time"
	"flag"
	"github.com/tidwall/gjson"
	"golang.org/x/net/webdav"
//    "context"
)

var servicePort int64
var prefixDir string
var ffilename = flag.String("f", "config.json", "配置文件名称")
var configFileName string
var readFileTicker = time.NewTicker(10 * time.Second)
var chJsonStr = make(chan string)

func init() {

	flag.Parse()
	configFileName = *ffilename
}

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

	//用信道对多协程读写配置文件资源进行同步----------
	jsonData := <-chJsonStr

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

        //下面处理让浏览器也可以在线查看文件(原来浏览器无法查看webdav文件)=====
        if req.Method == "GET" && handleDirList(prefixDir, w, req) {
            return
        }
        //=====================================================================

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
	jsonData, err := getStringFromFile(configFileName)
	if err != nil {
		log.Println("读配置文件失败，请确认已配置好文件config.json")
		return
	}
	servicePort = gjson.Get(jsonData, "serviceport").Int()
	prefixDir = gjson.Get(jsonData, "prefixdir").String()

	s_mux := http.NewServeMux()
	s_mux.HandleFunc("/", httpHandler)

	dav_addr := fmt.Sprintf(":%v", servicePort)
	log.Println("webDav Server run port", dav_addr)

	//读ca证书的公钥、私钥，如读成功，启动https服务，如果读不成功启动http服务
	caKey := gjson.Get(jsonData, "cakey")
	caCrt := gjson.Get(jsonData, "cacrt")

	if !caKey.Exists() || !caCrt.Exists() {
		log.Println("webDav server run http mode")
		//http.ListenAndServe是阻塞语句，不出错，后面的语句不会执行
		err := http.ListenAndServe(dav_addr, s_mux)
		if err != nil {
			log.Println("webDav server run http mod error:", err)
		}
	} else if !checkFileIsExist(caKey.String()) || !checkFileIsExist(caCrt.String()){
 		log.Println("webDav server run http mode")
		//http.ListenAndServe是阻塞语句，不出错，后面的语句不会执行
		err := http.ListenAndServe(dav_addr, s_mux)
		if err != nil {
			log.Println("webDav server run http mod error:", err)
		}
   
    
    
    } else {
        log.Println("caName is: ",caKey.String(), caCrt.String())
		log.Println("webDav server run  https mode")
		//http.ListenAndServeTLS是阻塞语句，不出错，后面的语句不会执行
		err := http.ListenAndServeTLS(dav_addr, caCrt.String(), caKey.String(), s_mux)
		if err != nil {
			log.Println("webDav server run https mod error:", err)
		}
	}

}

func main() {
	/*
	   1. 用一个go协程开启管理端口服务(具体服务程序放后期开发完善)
	   2. 开启webdav服务
	*/
	go FanInJsonStr()
	webDavLoad()

}

//通过信道为httpHandler函数提供配置文件数据
//不能使用带缓冲的信道，否则配置文件已经修改变新了，还得等一些请求取走信道中的旧数据使用
func FanInJsonStr() {

	jsonStr, _ := getStringFromFile(configFileName)
	for {
		select {
		case <-readFileTicker.C:
			jsonStr, _ = getStringFromFile(configFileName)
			chJsonStr <- jsonStr
		default:
			chJsonStr <- jsonStr
		}
	}

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

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func handleDirList(prefixDir string,  w http.ResponseWriter, req *http.Request) bool {
    f, err := os.OpenFile( prefixDir + req.URL.Path, os.O_RDONLY, 0)
    if err != nil {
        log.Println("dir Open file err:=",err,req.URL.Path)
        return false
    }
    defer f.Close()
    if fi, _ := f.Stat(); fi != nil && !fi.IsDir() {
        return false
    }
    dirs, err := f.Readdir(-1)
    if err != nil {
        log.Print(w, "Error reading directory", http.StatusInternalServerError)
        return false
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    fmt.Fprintf(w, "<pre>\n")
    for _, d := range dirs {
        name := d.Name()
        if d.IsDir() {
            name += "/"
        }
        fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", name, name)
    }
    fmt.Fprintf(w, "</pre>\n")
    return true
}
