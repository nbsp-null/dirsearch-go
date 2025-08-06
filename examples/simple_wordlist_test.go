package main

import (
	"fmt"
	"log"
	"time"

	"dirsearch-go/internal/api"
)

func main() {
	fmt.Println("=== 简单URL Wordlist测试 ===")

	// 测试基本功能
	fmt.Println("\n测试URL wordlist加载...")

	start := time.Now()

	// 使用一个简单的wordlist URL进行测试
	results, err := api.ScanSingleURLWithWordlist(
		"https://httpbin.org",
		"https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt",
		[]int{200, 403},
	)

	duration := time.Since(start)

	if err != nil {
		log.Printf("测试失败: %v", err)
		return
	}

	fmt.Printf("✅ URL wordlist功能测试成功!\n")
	fmt.Printf("📊 扫描统计:\n")
	fmt.Printf("  - 扫描时间: %v\n", duration)
	fmt.Printf("  - 发现结果: %d 个\n", len(results))

	if len(results) > 0 {
		fmt.Printf("  - 前3个结果:\n")
		for i, result := range results {
			if i >= 3 {
				break
			}
			fmt.Printf("    [%d] %s\n", result.StatusCode, result.URL)
		}
	}

	fmt.Println("\n🎉 URL wordlist API功能正常工作!")
}
