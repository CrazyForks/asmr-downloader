package spider

import (
	"asmr-downloader/config"
	"asmr-downloader/model"
	"asmr-downloader/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/xxjwxc/gowp/workpool"
	"io"
	"log"
	"net/http"
	"strconv"
)

var ctx = context.Background()

// ASMRClient ASMR 客户端
type ASMRClient struct {
	GlobalConfig  *config.Config
	Authorization string
	WorkerPool    *workpool.WorkPool
}

// 音轨
type track struct {
	Type             string  `json:"type"`
	Title            string  `json:"title"`
	Children         []track `json:"children,omitempty"`
	Hash             string  `json:"hash,omitempty"`
	WorkTitle        string  `json:"workTitle,omitempty"`
	MediaStreamURL   string  `json:"mediaStreamUrl,omitempty"`
	MediaDownloadURL string  `json:"mediaDownloadUrl,omitempty"`
}

// NewASMRClient 初始化ASMR客户端
func NewASMRClient(maxWorker int, globalConfig *config.Config) *ASMRClient {
	return &ASMRClient{
		WorkerPool:   utils.NewWorkerPool(maxWorker),
		GlobalConfig: globalConfig,
	}
}

// Login 登入获取授权信息
func (ac *ASMRClient) Login() error {
	payload, err := json.Marshal(map[string]string{
		"name":     ac.GlobalConfig.Account,
		"password": ac.GlobalConfig.Password,
	})
	if err != nil {
		fmt.Println("登录失败, 配置文件有误。")
		return err
	}
	client := utils.Client.Get().(*http.Client)
	req, _ := http.NewRequest("POST", "https://api.asmr.one/api/auth/me", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://www.asmr.one/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	resp, err := client.Do(req)
	utils.Client.Put(client)
	if err != nil {
		fmt.Println("登录失败, 网络错误。请尝试通过环境变量的方式设置代理。")
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("登录失败, 读取响应失败。")
		return err
	}
	res := make(map[string]string)
	err = json.Unmarshal(all, &res)
	ac.Authorization = "Bearer " + res["token"]
	return nil
}

func GetIndexPageInfo(authorStr string) (*model.PageResult, error) {
	var seed int = utils.GenerateReqSeed()
	randomUserAgent := browser.Random()
	log.Printf("Random: %s\n", randomUserAgent)
	var reqUrl = "https://api.asmr.one/api/works?order=create_date&sort=desc&page=1&seed=" + strconv.Itoa(seed) + "&subtitle=0"

	var resp = new(model.PageResult)
	client := &http.Client{}
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		// Handle error
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	//req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "zh,en;q=0.9,zh-TW;q=0.8,zh-CN;q=0.7,ja;q=0.6")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2FzbXIub25lIiwic3ViIjoicGV0ZXJsaXUiLCJhdWQiOiJodHRwczovL2FzbXIub25lL2FwaSIsIm5hbWUiOiJwZXRlcmxpdSIsImdyb3VwIjoidXNlciIsImlhdCI6MTY3NTYxOTc4MiwiZXhwIjoxNzA3MTU1NzgyfQ.OF5PIjC9G024-_00ujujj8-y1NXfSWOtkOGWOln_XRA")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Origin", "https://www.asmr.one")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://www.asmr.one/")
	req.Header.Set("Sec-Ch-UA", `"Not?A_Brand";v="8", "Chromium";v="108", "Google Chrome";v="108"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "macOS")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	respond, respError := client.Do(req.WithContext(context.Background()))
	utils.Client.Put(client)

	if respError != nil {
		fmt.Println("请求失败: ", respError.Error())
		return nil, respError
	}
	defer func() { _ = respond.Body.Close() }()
	all, err := io.ReadAll(respond.Body)
	if err != nil {
		fmt.Println("获取首页信息失败: ", err)
		return nil, err
	}
	err = json.Unmarshal(all, resp)
	return resp, nil
}
