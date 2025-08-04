package connection

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"dirsearch-go/internal/config"
)

// Response HTTP响应
type Response struct {
	StatusCode    int
	ContentLength int64
	Body          string
	Redirect      string
	Headers       http.Header
}

// Requester HTTP请求器
type Requester struct {
	client      *http.Client
	config      *config.Config
	headers     map[string]string
	HostManager *HostManager
}

// NewRequester 创建新的请求器
func NewRequester(cfg *config.Config) (*Requester, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("NewRequester panic recovered: %v\nStack trace: %s", r, debug.Stack())
		}
	}()

	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(cfg.Connection.Timeout) * time.Second,
	}

	// 设置代理
	if cfg.Connection.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Connection.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	// 设置请求头
	headers := make(map[string]string)
	if cfg.Request.UserAgent != "" {
		headers["User-Agent"] = cfg.Request.UserAgent
	} else {
		headers["User-Agent"] = "DirSearch-Go/1.0"
	}

	// 添加自定义请求头
	for _, header := range cfg.Request.Headers {
		if strings.Contains(header, ":") {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// 设置Cookie
	if cfg.Request.Cookie != "" {
		headers["Cookie"] = cfg.Request.Cookie
	}

	return &Requester{
		client:      client,
		config:      cfg,
		headers:     headers,
		HostManager: NewHostManager(cfg),
	}, nil
}

// Request 发送HTTP请求
func (r *Requester) Request(targetURL string) (*Response, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Request panic recovered: %v\nStack trace: %s", r, debug.Stack())
		}
	}()

	// 解析URL获取主机名
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// 获取主机信息（包含ping延迟，自动进行ping验证）
	r.HostManager.GetOrCreateHostInfo(parsedURL.Host)

	// 创建请求
	var req *http.Request
	method := strings.ToUpper(r.config.Request.HTTPMethod)

	if method == "POST" || method == "PUT" || method == "PATCH" {
		var body io.Reader
		if r.config.Request.Data != "" {
			body = strings.NewReader(r.config.Request.Data)
		}
		req, err = http.NewRequest(method, targetURL, body)
	} else {
		req, err = http.NewRequest(method, targetURL, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	for key, value := range r.headers {
		req.Header.Set(key, value)
	}

	// 设置认证
	if r.config.Request.Auth != "" {
		if r.config.Request.AuthType == "basic" {
			req.SetBasicAuth("", r.config.Request.Auth)
		} else if r.config.Request.AuthType == "bearer" {
			req.Header.Set("Authorization", "Bearer "+r.config.Request.Auth)
		}
	}

	// 设置智能超时
	timeout := r.HostManager.GetTimeout(parsedURL.Host)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req = req.WithContext(ctx)

	// 记录请求开始时间
	startTime := time.Now()

	// 发送请求
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	// 计算响应时间
	responseTime := time.Since(startTime)

	// 判断是否为慢响应
	isSlowResponse := r.HostManager.IsSlowResponse(parsedURL.Host, responseTime)

	// 读取响应体（根据响应速度决定是否完整读取）
	var bodyBytes []byte
	if isSlowResponse {
		// 慢响应：只读取前1KB用于状态码判断
		bodyBytes = make([]byte, 1024)
		n, _ := io.ReadAtLeast(resp.Body, bodyBytes, 1)
		bodyBytes = bodyBytes[:n]
	} else {
		// 正常响应：完整读取
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	}

	// 处理重定向
	var redirect string
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		if location := resp.Header.Get("Location"); location != "" {
			redirect = location
		}
	}

	return &Response{
		StatusCode:    resp.StatusCode,
		ContentLength: int64(len(bodyBytes)),
		Body:          string(bodyBytes),
		Redirect:      redirect,
		Headers:       resp.Header,
	}, nil
}

// SetHeaders 设置请求头
func (r *Requester) SetHeaders(headers map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("SetHeaders panic recovered: %v", r)
		}
	}()

	if headers != nil {
		for key, value := range headers {
			r.headers[key] = value
		}
	}
}

// SetUserAgent 设置用户代理
func (r *Requester) SetUserAgent(userAgent string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("SetUserAgent panic recovered: %v", r)
		}
	}()

	if userAgent != "" {
		r.headers["User-Agent"] = userAgent
	}
}

// SetCookie 设置Cookie
func (r *Requester) SetCookie(cookie string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("SetCookie panic recovered: %v", r)
		}
	}()

	if cookie != "" {
		r.headers["Cookie"] = cookie
	}
}
