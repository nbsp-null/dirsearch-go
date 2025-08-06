package main

import (
	"fmt"
	"log"
	"time"

	"dirsearch-go/internal/api"
)

func main() {
	fmt.Println("=== DirSearch-Go URL Wordlist API 示例 ===")

	// 示例1: 基本URL wordlist扫描
	fmt.Println("\n1. 基本URL wordlist扫描")
	basicScan()

	// 示例2: 高级URL wordlist扫描
	fmt.Println("\n2. 高级URL wordlist扫描")
	advancedScan()

	// 示例3: 自定义状态码过滤
	fmt.Println("\n3. 自定义状态码过滤")
	customStatusFilter()
}

// basicScan 基本URL wordlist扫描示例
func basicScan() {
	start := time.Now()

	results, err := api.ScanSingleURLWithWordlist(
		"https://httpbin.org",
		"https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt",
		[]int{200, 403}, // 只显示200和403状态码
	)
	if err != nil {
		log.Printf("基本扫描失败: %v", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("扫描完成，用时: %v\n", duration)
	fmt.Printf("发现 %d 个结果:\n", len(results))

	for i, result := range results {
		if i >= 5 { // 只显示前5个结果
			fmt.Printf("... 还有 %d 个结果\n", len(results)-5)
			break
		}
		fmt.Printf("  [%d] %s\n", result.StatusCode, result.URL)
	}
}

// advancedScan 高级URL wordlist扫描示例
func advancedScan() {
	start := time.Now()

	// 创建高级扫描选项
	options := &api.ScanOptions{
		Threads:       5,    // 减少线程数
		Timeout:       5.0,  // 设置超时
		ShowAllStatus: true, // 显示所有状态码
		UserAgent:     "DirSearch-Go-API/1.0",
		RecursiveScan: false, // 禁用递归扫描
	}

	response, err := api.ScanSingleURLWithWordlistAdvanced(
		"https://httpbin.org",
		"https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt",
		options,
	)
	if err != nil {
		log.Printf("高级扫描失败: %v", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("高级扫描完成，用时: %v\n", duration)
	fmt.Printf("总扫描数: %d, 发现: %d, 错误: %d\n",
		response.TotalScanned, response.TotalFound, response.TotalErrors)

	// 显示状态码统计
	fmt.Println("状态码分布:")
	for statusCode, count := range response.StatusSummary {
		fmt.Printf("  %d: %d\n", statusCode, count)
	}

	// 显示前几个结果
	fmt.Println("扫描结果:")
	for i, result := range response.Results {
		if i >= 3 { // 只显示前3个结果
			fmt.Printf("... 还有 %d 个结果\n", len(response.Results)-3)
			break
		}
		fmt.Printf("  [%d] %s (长度: %d)\n",
			result.StatusCode, result.URL, result.ContentLength)
	}
}

// customStatusFilter 自定义状态码过滤示例
func customStatusFilter() {
	start := time.Now()

	// 只扫描特定状态码
	statusCodes := []int{200, 301, 302, 403, 404}

	results, err := api.ScanSingleURLWithWordlist(
		"https://httpbin.org",
		"https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt",
		statusCodes,
	)
	if err != nil {
		log.Printf("自定义状态码扫描失败: %v", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("自定义状态码扫描完成，用时: %v\n", duration)
	fmt.Printf("发现 %d 个结果 (状态码: %v):\n", len(results), statusCodes)

	// 按状态码分组显示
	statusGroups := make(map[int][]api.ScanResult)
	for _, result := range results {
		statusGroups[result.StatusCode] = append(statusGroups[result.StatusCode], result)
	}

	for statusCode, group := range statusGroups {
		fmt.Printf("  状态码 %d (%d 个):\n", statusCode, len(group))
		for i, result := range group {
			if i >= 2 { // 每个状态码只显示前2个
				fmt.Printf("    ... 还有 %d 个\n", len(group)-2)
				break
			}
			fmt.Printf("    %s\n", result.URL)
		}
	}
}

// 辅助函数：格式化时间
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
