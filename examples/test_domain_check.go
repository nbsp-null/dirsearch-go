package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func main() {
	fmt.Println("=== 域名检测测试 ===")

	// 测试域名列表
	testURLs := []string{
		"https://httpbin.org",
		"https://example.com",
		"https://google.com",
		"https://github.com",
	}

	for _, testURL := range testURLs {
		fmt.Printf("\n测试域名: %s\n", testURL)

		// 解析URL
		parsedURL, err := url.Parse(testURL)
		if err != nil {
			fmt.Printf("  ❌ URL解析失败: %v\n", err)
			continue
		}

		// 构建检测URL
		checkURL := fmt.Sprintf("%s://%s/", parsedURL.Scheme, parsedURL.Host)
		fmt.Printf("  检测URL: %s\n", checkURL)

		// 创建HTTP客户端
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		// 创建请求
		req, err := http.NewRequest("HEAD", checkURL, nil)
		if err != nil {
			fmt.Printf("  ❌ 创建请求失败: %v\n", err)
			continue
		}

		// 设置正常的浏览器请求头
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Upgrade-Insecure-Requests", "1")

		// 发送请求
		start := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("  ❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		// 检查结果
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			fmt.Printf("  ✅ 域名存活 (状态码: %d, 用时: %v)\n", resp.StatusCode, duration)
		} else {
			fmt.Printf("  ❌ 域名不存活 (状态码: %d, 用时: %v)\n", resp.StatusCode, duration)
		}
	}
}
