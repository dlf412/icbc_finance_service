package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"github.com/kardianos/service"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func IntPtr(n int) uintptr {
	return uintptr(n)
}

func StrPtr(s string) uintptr {
	return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
}

// windows下的另一种DLL方法调用
func ShowMessage2(title, text string) {
	user32dll, _ := syscall.LoadLibrary("user32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")
	MessageBoxW := user32.NewProc("MessageBoxW")
	MessageBoxW.Call(IntPtr(0), StrPtr(text), StrPtr(title), IntPtr(0))
	defer syscall.FreeLibrary(user32dll)
}



func portInUse(portNumber int) int {
	res := -1
	var outBytes bytes.Buffer
	cmdStr := fmt.Sprintf("netstat -ano -p tcp | findstr %d", portNumber)
	cmd := exec.Command("cmd", "/c", cmdStr)
	cmd.Stdout = &outBytes
	cmd.Run()
	resStr := outBytes.String()
	r := regexp.MustCompile(`\s\d+\s`).FindAllString(resStr, -1)
	if len(r) > 0 {
		pid, err := strconv.Atoi(strings.TrimSpace(r[0]))
		if err != nil {
			res = -1
		} else {
			res = pid
		}
	}
	return res
}


type Res struct {
	Url string `json:"url"`
}

var save_dir = os.TempDir()
var addr = ""
var serviceConfig = &service.Config{
	Name:        "ICBCZLDFinance",
	DisplayName: "ICBC_ZLD FINANCE SERVICE",
	Description: "工商银行企业财务室代发工资本地服务",
	Option: service.KeyValue{"DelayedAutoStart": true},

}

func main() {

	// 构建服务对象
	prog := &Program{}
	serviceConfig.Arguments = []string{os.TempDir()}
	s, err := service.New(prog, serviceConfig)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) < 2 {
		err = s.Install()
		if err != nil {
			err_msg := fmt.Sprintf("%s", err)
			if !strings.Contains(err_msg, "already exists") {
				ShowMessage2("安装错误提示", err_msg)
				log.Fatal(err)
			}
		}

		err = s.Start()
		if err != nil {
			err_msg := fmt.Sprintf("%s", err)
			if strings.Contains(err_msg, "is already running"){
				err_msg = "服务实例正在运行中，请勿重复运行！"
			}
			ShowMessage2("启动错误提示", err_msg)
			log.Fatal(err)
		}

		ShowMessage2("系统提示", fmt.Sprintf("服务启动成功"))
		return
	}

	cmd := os.Args[1]

	if cmd == "install" {
		err = s.Install()
		if err != nil {
			ShowMessage2("系统提示", fmt.Sprintf("%s", err))
			log.Fatal(err)
		}
		//fmt.Println("服务安装成功, 开始启动服务")
		err = s.Start()
		if err != nil {
			ShowMessage2("系统提示", fmt.Sprintf("%s", err))
			log.Fatal(err)
		}
		ShowMessage2("系统提示", fmt.Sprintf("服务启动成功"))
		return
	}
	if cmd == "uninstall" {
		s.Stop()
		err = s.Uninstall()
		if err != nil {
			ShowMessage2("系统提示", fmt.Sprintf("%s", err))
			log.Fatal(err)
		}
		ShowMessage2("系统提示", fmt.Sprintf("服务卸载成功"))
	} else {   // 服务启动的时候带参数，参数为文件保存的临时目录
		save_dir = cmd
		err = s.Run()
		if err != nil {
			log.Fatal(err)
		}
		return
	}
}

type Program struct{}

func (p *Program) Start(s service.Service) error {
	for port:=8654; port< 8900; port += 100 {
		pid := portInUse(port)
		if pid == -1 {
			addr = fmt.Sprintf("localhost:%d", port)
			break
		}
	}
	if addr == "" {
		return errors.New("服务启动失败，端口被占用")
	}
	// addr := fmt.Sprintf("localhost:%d", port)
	go p.run(addr)
	return nil
}

func (p *Program) Stop(s service.Service) error {
	log.Println("停止服务")
	return nil
}

func (p *Program) run(addr string) {
	http.HandleFunc("/icbc_payroll", index)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func download(url, filename string) {
	// 下载url的文件，并保存成临时文件, 返回该临时文件路径
	r, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer func() {_ = r.Body.Close()}()

	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer func() {_ = f.Close()}()

	n, err := io.Copy(f, r.Body)
	fmt.Println(n, err)
}

func httpReq(client *http.Client, url string, sid string) (Res, error){
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// handle error
		return Res{}, err
	}
	req.Header.Set("Authorization", "SID " + sid)
	resp, err := client.Do(req)
	if err != nil {
		return Res{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Res{}, errors.New("请求响应码出错:" + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		return Res{}, err
	}

	res := Res{}
	json.Unmarshal(body, &res)

	return res, nil
}

func httpDo(domain string, sid string, payroll string) (string, error) {

	client := &http.Client{}
	res, err := httpReq(client, domain + "/api/v1/icbc_payroll?payroll_id=" + payroll + "&method=download", sid)
	if err != nil {
		// handle error
		return "", err
	}
	download_url := domain + res.Url
	// dir := os.TempDir()
	filename := save_dir + string(os.PathSeparator) + "工资单_" + payroll + ".xlsx"
	download(download_url, filename)
	// 请求url
	res, err = httpReq(client, domain + "/api/v1/icbc_payroll?payroll_id=" + payroll + "&method=pay&args=" + filename, sid)
	return res.Url, nil
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json

	values := r.URL.Query()
	fmt.Println(values) // map[string][]string
	domain := values["domain"][0]
	sid := values["sid"][0]
	payroll := values["payroll_id"][0]

	// 收到请求，转发给后端服务器，获取资源文件，并下载
	pay_url, err := httpDo(domain, sid, payroll)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"detail": "%s"}`, err)
		return
	}
	fmt.Fprintf(w, `{"url":"%s"}`, pay_url)
}