package main

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

const (
	// 熔断器的名字，一个名字对应一个熔断器
	commandName = "my_command"
)

type ResponseWriter interface {
	Header() http.Header

	Write([]byte) (int, error)

	WriteHeader(statusCode int)
}

type CusomResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *CusomResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *CusomResponseWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

func (w *CusomResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

type HystrixHandler struct {
	next    http.Handler
	command hystrix.CommandConfig
}

func NewHystrixHandler(next http.Handler, config hystrix.CommandConfig) *HystrixHandler {
	hystrix.ConfigureCommand(
		commandName, // 熔断器的名字，一个名字对应一个熔断器
		config,
	)
	return &HystrixHandler{
		next:    next,
		command: config,
	}
}

func (h *HystrixHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hystrix.DoC(r.Context(), commandName, func(ctx context.Context) error {
		//包装原始的 http.ResponseWriter
		stastusCodeRw := &CusomResponseWriter{ResponseWriter: w}
		h.next.ServeHTTP(stastusCodeRw, r)
		if stastusCodeRw.StatusCode >= http.StatusMultipleChoices {
			// 服务返回状态码大于等于 300，返回错误
			return hystrix.CircuitError{
				Message: "服务返回错误",
			}
		}

		return nil
	}, func(ctx context.Context, err error) error {
		// 服务暂时熔断，返回状态码 503
		w.WriteHeader(http.StatusServiceUnavailable)
		return nil
	})
}

// 普通的 http.Handler
type NormalHandler struct{}

func (h *NormalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 返回hello world
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(10) > 5 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello world"))
}

func main() {
	// 设置熔断器
	hyConfig := hystrix.CommandConfig{
		Timeout:                1000, //超时时间，单位毫秒 ms。默认 1000ms
		MaxConcurrentRequests:  2,    // 最大并发数，超过这个设置就返回错误。默认 10
		ErrorPercentThreshold:  50,   // 设置错误数量统计百分比阙值，超过这个阙值，就开启熔断。默认 50
		RequestVolumeThreshold: 4,    // 一个窗口10秒内请求(有问题的请求)的数量阙值，达到这个阙值就开启熔断
		SleepWindow:            5000, // 熔断器被激活后，多久重试服务是否可用，单位毫秒。默认 5000ms
	}

	// 设置熔断器
	hyHandler := NewHystrixHandler(&NormalHandler{}, hyConfig)

	// 启动熔断器监控
	registerHandler()

	// 注册路由
	http.Handle("/hello", hyHandler)
	http.ListenAndServe(":80", nil)
}

func registerHandler() {
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	http.Handle("/", hystrixStreamHandler)
}
