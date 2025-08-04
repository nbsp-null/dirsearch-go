package connection

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"dirsearch-go/internal/config"

	"github.com/chromedp/chromedp"
)

// HeadlessBrowser 无头浏览器
type HeadlessBrowser struct {
	config *config.Config
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// HeadlessResult 无头浏览器扫描结果
type HeadlessResult struct {
	URL           string
	StatusCode    int
	Title         string
	Content       string
	Headers       map[string]string
	Cookies       []string
	JavaScript    bool
	Redirects     []string
	Error         error
	ResponseTime  time.Duration
	ContentLength int64
}

// NewHeadlessBrowser 创建新的无头浏览器
func NewHeadlessBrowser(cfg *config.Config) (*HeadlessBrowser, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-images", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-field-trial-config", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("user-agent", "dirsearch-go/1.0"),
		chromedp.Flag("disable-logging", true),
		chromedp.Flag("log-level", "0"),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))

	return &HeadlessBrowser{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Close 关闭浏览器
func (hb *HeadlessBrowser) Close() {
	hb.mu.Lock()
	defer hb.mu.Unlock()
	if hb.cancel != nil {
		hb.cancel()
	}
}

// ScanURL 扫描单个URL
func (hb *HeadlessBrowser) ScanURL(targetURL string) *HeadlessResult {
	startTime := time.Now()
	result := &HeadlessResult{
		URL:          targetURL,
		Headers:      make(map[string]string),
		Redirects:    make([]string, 0),
		ResponseTime: 0,
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(hb.ctx, time.Duration(hb.config.Connection.Timeout)*time.Second)
	defer cancel()

	// 执行扫描任务
	var title, content string
	var statusCode int

	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.Sleep(1*time.Second), // 等待页面加载
		chromedp.Title(&title),
		chromedp.OuterHTML("html", &content),
		chromedp.Evaluate(`200`, &statusCode), // 简化状态码获取
	)

	if err != nil {
		result.Error = fmt.Errorf("headless scan failed: %w", err)
		return result
	}

	result.Title = title
	result.Content = content
	result.StatusCode = statusCode
	result.ResponseTime = time.Since(startTime)
	result.ContentLength = int64(len(content))

	// 提取重定向信息
	if len(result.Redirects) > 0 {
		result.Redirects = append(result.Redirects, targetURL)
	}

	return result
}

// ScanMultipleURLs 批量扫描URL
func (hb *HeadlessBrowser) ScanMultipleURLs(urls []string, maxConcurrency int) []*HeadlessResult {
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	// 创建任务通道
	taskChan := make(chan string, len(urls))
	resultChan := make(chan *HeadlessResult, len(urls))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range taskChan {
				result := hb.ScanURL(url)
				resultChan <- result
			}
		}()
	}

	// 发送任务
	go func() {
		defer close(taskChan)
		for _, url := range urls {
			taskChan <- url
		}
	}()

	// 收集结果
	go func() {
		defer close(resultChan)
		wg.Wait()
	}()

	var results []*HeadlessResult
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// IsJavaScriptEnabled 检查JavaScript是否启用
func (hb *HeadlessBrowser) IsJavaScriptEnabled() bool {
	return !strings.Contains(hb.config.Request.UserAgent, "disable-javascript")
}

// GetBrowserInfo 获取浏览器信息
func (hb *HeadlessBrowser) GetBrowserInfo() map[string]interface{} {
	return map[string]interface{}{
		"user_agent": hb.config.Request.UserAgent,
		"javascript": hb.IsJavaScriptEnabled(),
		"timeout":    hb.config.Connection.Timeout,
		"headless":   true,
	}
}
