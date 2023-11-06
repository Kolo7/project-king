package main

// export GOOS=darwin GOARCH=arm64 && go build -o script main.go

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"git.imgo.tv/ft/hystrix-go/hystrix"
	"github.com/pkg/errors"
)

const (
	defaultSpeed     = 3
	defaultTokenUrl  = "https://www.infocomm-journal.com/dxkx/CN/article/showArticleFile.do"
	defaultActionUrl = "https://www.infocomm-journal.com/dxkx/CN/article/downloadArticleFileFee.do"
	actionCommand    = "action"
	tokenCommand     = "token"
)

var (
	speed     int    // 请求速率，每秒请求数
	tokenUrl  string // 获取token的url
	actionUrl string // 请求的url
	ticketCh  chan struct{}
)

func main() {
	// 读取命令行参数 -s
	flag.Parse()
	ctx := context.Background()
	for {
		<-ticketCh // 限速
		// 读取token
		var token string
		err := hystrix.DoC(ctx, tokenCommand, func(ctx context.Context) (err error) {
			token, err = getToken(tokenUrl)
			if err != nil {
				return err
			}
			return
		}, func(ctx context.Context, err error) error {
			log.Printf("熔断中, 错误: %v\n", err)
			return nil
		})
		if err != nil {
			log.Panicf("获取 token 失败, err: %v\n", err)
		}
		err = hystrix.DoC(ctx, tokenCommand, func(ctx context.Context) (err error) {
			err = action(actionUrl, token)
			if err != nil {
				return err
			}
			return
		}, func(ctx context.Context, err error) error {
			log.Printf("熔断中,错误: %v\n", err)
			return nil
		})
		if err != nil {
			log.Panicf("请求访问失败, err: %v\n", err)
		}
	}

}

func action(query, token string) (err error) {
	body := fmt.Sprintf("attachType=RICH_HTML&id=159681&json=true&token=%s&referer=", token)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s?%d", query, time.Now().UnixMilli()), bytes.NewBufferString(body))
	if err != nil {
		err = errors.Wrapf(err, "请求访问失败: %s", query)
		return
	}
	// 设置请求头
	setHeader(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = errors.Wrapf(err, "请求访问失败, url: %s", query)
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.Errorf("请求访问失败, status code: %d", resp.StatusCode)
		return
	}
	return
}

func getToken(query string) (token string, err error) {
	urlQuery := fmt.Sprintf("https://www.infocomm-journal.com/dxkx/CN/article/showArticleFile.do?1698650984316=null")
	payload := strings.NewReader(`attachType=RICH_HTML&id=159681&json=true`)
	req, err := http.NewRequest(http.MethodPost, urlQuery, payload)
	if err != nil {
		err = errors.Wrapf(err, "获取 token 失败, url: %s", query)
		return
	}

	// 设置请求头
	setHeader(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = errors.Wrapf(err, "获取 token 失败, url: %s", query)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = errors.Errorf("获取 token 失败, status code: %d", resp.StatusCode)
		return
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	token, err = handleResp(string(respBytes))
	if err != nil {
		err = fmt.Errorf("获取 token 失败, err: %v", err)
		return
	}

	return
}

// 设置请求头
func setHeader(req *http.Request) {
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "zh,en;q=0.9,zh-CN;q=0.8,en-US;q=0.7")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Cookie", "acw_tc=276aedcf16986301909836506e796724251d3077b8569082faf4935b791a81; wkxt3_csrf_token=b91951e1-4ce8-4147-a0f2-525310f994cd; JSESSIONID=AEDD2F5287D1A477D7410753E98F8768")
	req.Header.Add("Origin", "https://www.infocomm-journal.com")
	req.Header.Add("Referer", "https://www.infocomm-journal.com/dxkx/CN/10.11959/j.issn.1000-0801.2015322")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("sec-ch-ua", "\"Chromium\";v=\"118\", \"Google Chrome\";v=\"118\", \"Not=A?Brand\";v=\"99\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"macOS\"")
}

func handleResp(str string) (string, error) {
	if len(str) < 6 {
		return "", fmt.Errorf("str len < 6")
	}
	str = str[6:]
	// 通过json解析，转为map
	obj := make(map[string]interface{})
	err := json.Unmarshal([]byte(str), &obj)
	if err != nil {
		return "", err
	}
	if _, ok := obj["clickRichToken"]; !ok {
		return "", fmt.Errorf("clickRichToken not exist")
	}
	if _, ok := obj["clickRichToken"].(string); !ok {
		return "", fmt.Errorf("clickRichToken not string")
	}
	return obj["clickRichToken"].(string), nil
}

func init() {
	// 帮助信息
	flag.Usage = func() {
		fmt.Println("Usage: ./script -s 3 -t https://www.infocomm-journal.com/dxkx/CN/article/showArticleFile.do -a https://www.infocomm-journal.com/dxkx/CN/article/downloadArticleFileFee.do")
		fmt.Println("Options:")
		fmt.Println("  -h\t\t帮助信息")
		fmt.Println("  -s\t\t请求速率，每秒请求数")
		fmt.Println("  -t\t\t获取token的url")
		fmt.Println("  -a\t\t请求的url")
		fmt.Println("  -timeout\t熔断超时时间，单位毫秒 ms。默认 1000ms")
		fmt.Println("  -errorPercentThreshold\t错误数量统计百分比阙值，超过这个阙值，就开启熔断。默认 50")
		fmt.Println("  -volumeThreshold\t一个窗口10秒内请求(有问题的请求)的数量阙值，达到这个阙值就开启熔断")
		flag.PrintDefaults()
	}
	// -h 帮助信息
	help := flag.Bool("h", false, "help")
	if help != nil {
		flag.Usage()
	}
	flag.IntVar(&speed, "s", defaultSpeed, "speed")
	flag.StringVar(&tokenUrl, "t", defaultActionUrl, "token url")
	flag.StringVar(&actionUrl, "a", defaultActionUrl, "action url")
	// 熔断超时时间
	flag.IntVar(&hystrix.DefaultTimeout, "timeout", 1000, "timeout")
	// 错误数量统计百分比阙值
	flag.IntVar(&hystrix.DefaultErrorPercentThreshold, "errorPercentThreshold", 50, "errorPercentThreshold")
	// 一个窗口10秒内请求(有问题的请求)的数量阙值，达到这个阙值就开启熔断
	flag.IntVar(&hystrix.DefaultVolumeThreshold, "volumeThreshold", 4, "volumeThreshold")
	// 熔断器被激活后，多久重试服务是否可用，单位毫秒
	if speed <= 0 || speed > 100 {
		speed = defaultSpeed
	}
	// 打印当前配置
	log.Printf("当前请求速率: %d\n", speed)
	log.Printf("当前获取 token 的 url: %s\n", tokenUrl)
	log.Printf("当前请求的 url: %s\n", actionUrl)

	ticketCh = make(chan struct{}, 10)
	go do()

	// 设置熔断器
	hyConfig := hystrix.CommandConfig{
		Timeout:                1000, //超时时间，单位毫秒 ms。默认 1000ms
		MaxConcurrentRequests:  2,    // 最大并发数，超过这个设置就返回错误。默认 10
		ErrorPercentThreshold:  50,   // 设置错误数量统计百分比阙值，超过这个阙值，就开启熔断。默认 50
		RequestVolumeThreshold: 4,    // 一个窗口10秒内请求(有问题的请求)的数量阙值，达到这个阙值就开启熔断
		SleepWindow:            5000, // 熔断器被激活后，多久重试服务是否可用，单位毫秒。默认 5000ms
	}

	hystrix.ConfigureCommand(
		actionCommand, // 熔断器的名字，一个名字对应一个熔断器
		hyConfig,
	)

	hystrix.ConfigureCommand(
		tokenCommand, // 熔断器的名字，一个名字对应一个熔断器
		hyConfig,
	)

	// 启动熔断器监控
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	http.Handle("/", hystrixStreamHandler)
	go http.ListenAndServe(":80", nil)
}

// 补充令牌
func do() {
	ticker := time.NewTicker(time.Second / time.Duration(speed))
	for {
		select {
		case <-ticker.C:
			ticketCh <- struct{}{}
		}
	}
}
