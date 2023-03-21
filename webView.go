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
    "github.com/tidwall/gjson"
    //    "context"
)


func webViewLoad() {
    /*
    1. 读服务端口和目录前缀
    2. 开启web服务
    */
    jsonData, err := getStringFromFile(configFileName)
    if err != nil {
        log.Println("读配置文件失败，请确认已配置好文件config.json")
        return
    }
    servicePort = gjson.Get(jsonData, "viewport").Int()

    s_mux := http.NewServeMux()
    s_mux.HandleFunc("/", httpHandlerWebView)

    dav_addr := fmt.Sprintf(":%v", servicePort)
    log.Println("view Server run port", dav_addr)

    //读ca证书的公钥、私钥，如读成功，启动https服务，如果读不成功启动http服务
    caKey := gjson.Get(jsonData, "cakey")
    caCrt := gjson.Get(jsonData, "cacrt")

    if !caKey.Exists() || !caCrt.Exists() {
        log.Println("view server run http mode")
        //http.ListenAndServe是阻塞语句，不出错，后面的语句不会执行
        err := http.ListenAndServe(dav_addr, s_mux)
        if err != nil {
            log.Println("view server run http mod error:", err)
        }
    } else if !checkFileIsExist(caKey.String()) || !checkFileIsExist(caCrt.String()){
        log.Println("view server run http mode")
        //http.ListenAndServe是阻塞语句，不出错，后面的语句不会执行
        err := http.ListenAndServe(dav_addr, s_mux)
        if err != nil {
            log.Println("view server run http mod error:", err)
        }



    } else {
        log.Println("caName is: ",caKey.String(), caCrt.String())
        log.Println("view server run  https mode")
        //http.ListenAndServeTLS是阻塞语句，不出错，后面的语句不会执行
        err := http.ListenAndServeTLS(dav_addr, caCrt.String(), caKey.String(), s_mux)
        if err != nil {
            log.Println("view server run https mod error:", err)
        }
    }

}


func httpHandlerWebView(w http.ResponseWriter, req *http.Request) {

    //fmt.Println(req)
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
        http.Error(w, "view: need authorized!", http.StatusUnauthorized)
        //log.Println("no user")
        return
    }

    //判断用户口令是否正确，口令可以直接存储，也可以以md5码存储，这里程序进行自动判断。
    //防止把md5码当明码进行对比，不允许密码长度为32个字符
    if len(password) == 32 {
        http.Error(w, "view: need authorized!", http.StatusUnauthorized)
        return
    }
    userPath := gjson.Get(user.String(), "#(password="+password+").userpath")
    if !userPath.Exists() {
        passtemp := md5.Sum([]byte(password))
        //不能直接md5.Sum([]byte(password))[:] ,会报slice of unaddressable value错误
        userPath = gjson.Get(user.String(), "#(password="+hex.EncodeToString(passtemp[:])+").userpath")
        if !userPath.Exists() {
            http.Error(w, "view: need authorized!", http.StatusUnauthorized)
            //log.Println("wrong password")
            return

        }
    }

    //下面处理让浏览器也可以在线查看文件(原来浏览器无法查看webdav文件)=====
    if req.Method == "GET" && handleDirListAndMdFile(prefixDir, w, req,userPath.String()) {
        return
    }
    //=====================================================================


    w.WriteHeader(404)
}

func handleDirListAndMdFile(prefixDir string,  w http.ResponseWriter, req *http.Request,userPath string) bool {
    //去除url路径(不含域名及端口号)的最前面一段用户名以备用真实路径名替换
    path := req.URL.Path[strings.Index(req.URL.Path,"/")+1:]
    startIndex:=strings.Index(path,"/")
    if startIndex<0 {
        return false
    }
    path = path[startIndex:]

    filePath := prefixDir + userPath + path //取得服务器上文件完整文件名

    f, err := os.OpenFile( filePath, os.O_RDONLY, 0)
    if err != nil {
        log.Println("dir Open file err:=",err,req.URL.Path)
        return false
    }
    defer f.Close()
    //如果是MD文件进行渲染后输出，如果是其它文件直接输出文件内容,如果是目录输出目录
    fmt.Println(req.Header)
    if !isDir(filePath) {
        if filePath[len(filePath)-2:]=="md" {
            //            fmt.Println(strings.Index( req.Header.Get("User-Agent"),"obsidian"))
            w.Header().Set("Content-Type", "text/html; charset=utf-8")

            fmt.Fprintf(w, `<!doctype html>
            <html>
            <head>
            <meta charset="utf-8"/>
            <script src="https://cdn.jsdelivr.net/npm/mermaid@9.4.3/dist/mermaid.min.js"></script>
            <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
            </head>
            <body>
            <div id="render"></div>
            <div id="md">`)
            txt,_:= ioutil.ReadAll(f)
            //fmt.Fprintf(w,string(txt)+"\n")
            w.Write(txt)
            fmt.Fprintf(w,"\n"+`</div>
            <script>
            document.getElementById('md').innerHTML =marked.parse(document.getElementById('md').innerHTML);

            document.getElementById('md').innerHTML=document.getElementById('md').innerHTML.replace(/<pre><code class="language-mermaid">([\S\s]+?)<\/code><\/pre>/g,'<div class="mermaid" id="test">\n$1</div>');

            //document.getElementById('md').innerHTML=document.getElementById('md').innerHTML.replace(/(?<=mermaid[^><]*>[^><]*)(&.+?;)(?=\|[^><]*<\/div>)|(?<=mermaid[^><]*>[^><]*)(&.+?;)(?=[^><]*<\/div>)/g,'>');


            document.getElementById('md').innerHTML=document.getElementById('md').innerHTML.replace(/&.{0,4}gt;/g,'>');
            document.getElementById('md').innerHTML=document.getElementById('md').innerHTML.replace(/&.{0,4}lt;/g,'<');

            mermaid.initialize({startOnLoad:true});
            </script>

            </body>
            </html>`)
            return true
        }


        txt,_:= ioutil.ReadAll(f)
        w.Write(txt)
        return true

    }

    //输出目录信息
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
        fmt.Fprintf(w, "<a href=\"%s\">%s</a></br>\n", name, name)
    }
    fmt.Fprintf(w, "</pre>\n")
    return true
}

func isDir(path string) bool {
    file,err:=os.Stat(path)
    if err!=nil {
        return false
    }
    if file.IsDir(){
        return true
    }
    return false
}
