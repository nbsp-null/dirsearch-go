package main

import (
	"encoding/json"
	"fmt"
	"log"

	"dirsearch-go/internal/api"
)

func main() {
	// 示例1: 快速扫描单个URL
	fmt.Println("=== 示例1: 快速扫描单个URL ===")
	results, err := api.ScanSingleURL(
		"https://httpbin.org",
		[]string{"wordlists/common.txt"},
		[]int{200, 403}, // 只显示200和403状态码
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("发现 %d 个结果:\n", len(results))
	for _, result := range results {
		fmt.Printf("[%d] %s%s\n", result.StatusCode, result.URL, result.Path)
	}

	// 示例2: 完整扫描选项
	fmt.Println("\n=== 示例2: 完整扫描选项 ===")
	options := api.ScanOptions{
		URLs:          []string{"https://httpbin.org"},
		Wordlists:     []string{"wordlists/common.txt"},
		Threads:       10,
		Delay:         0.1,
		ShowAllStatus: true,                 // 显示所有状态码
		StatusFilter:  []int{200, 403, 404}, // 过滤特定状态码
		UserAgent:     "DirSearch-Go/1.0",
		Headers:       []string{"X-Custom: test"},
		Timeout:       10.0,
		RecursiveScan: false,
	}

	response, err := api.Scan(options)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("扫描统计:\n")
	fmt.Printf("  总扫描数: %d\n", response.TotalScanned)
	fmt.Printf("  总发现数: %d\n", response.TotalFound)
	fmt.Printf("  总错误数: %d\n", response.TotalErrors)
	fmt.Printf("  状态码分布: %v\n", response.StatusSummary)

	// 示例3: 批量扫描多个URL
	fmt.Println("\n=== 示例3: 批量扫描多个URL ===")
	urls := []string{
		"https://httpbin.org",
		"https://example.com",
	}

	results, err = api.QuickScan(urls, []string{"wordlists/common.txt"}, []int{200})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("批量扫描发现 %d 个200状态码结果:\n", len(results))
	for _, result := range results {
		fmt.Printf("[%d] %s%s\n", result.StatusCode, result.URL, result.Path)
	}

	// 示例4: 输出JSON格式
	fmt.Println("\n=== 示例4: 输出JSON格式 ===")
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonData))
}
