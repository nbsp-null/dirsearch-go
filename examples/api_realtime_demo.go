package main

import (
	"fmt"
	"log"
	"time"

	"dirsearch-go/internal/api"
)

func main() {
	fmt.Println("=== DirSearch-Go API 实时状态显示测试 ===")

	// 测试1: 基本实时扫描
	fmt.Println("\n1. 基本实时扫描测试")
	testBasicRealtimeScan()

	// 测试2: 高级实时扫描
	fmt.Println("\n2. 高级实时扫描测试")
	testAdvancedRealtimeScan()

	// 测试3: 自定义配置实时扫描
	fmt.Println("\n3. 自定义配置实时扫描测试")
	testCustomRealtimeScan()
}

// testBasicRealtimeScan 基本实时扫描测试
func testBasicRealtimeScan() {
	start := time.Now()

	fmt.Println("开始基本实时扫描...")
	results, err := api.ScanSingleURLWithWordlist(
		"https://httpbin.org",
		"https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt",
		[]int{200, 403},
	)
	if err != nil {
		log.Printf("基本实时扫描失败: %v", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("基本实时扫描完成，用时: %v\n", duration)
	fmt.Printf("发现 %d 个结果:\n", len(results))

	// 显示前10个结果
	for i, result := range results {
		if i >= 10 {
			fmt.Printf("... 还有 %d 个结果\n", len(results)-10)
			break
		}
		fmt.Printf("  [%d] %s\n", result.StatusCode, result.URL)
	}
}

// testAdvancedRealtimeScan 高级实时扫描测试
func testAdvancedRealtimeScan() {
	start := time.Now()

	// 创建高级扫描选项
	options := &api.ScanOptions{
		URLs:           []string{"https://httpbin.org"},
		Wordlists:      []string{"https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt"},
		Threads:        10, // 减少线程数以便观察进度
		Timeout:        5.0,
		ShowAllStatus:  true, // 显示所有状态码
		RealTimeStatus: true, // 启用实时状态显示
		UserAgent:      "DirSearch-Go-API-Test/1.0",
	}

	fmt.Println("开始高级实时扫描...")
	response, err := api.Scan(*options)
	if err != nil {
		log.Printf("高级实时扫描失败: %v", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("高级实时扫描完成，用时: %v\n", duration)
	fmt.Printf("总扫描数: %d, 发现: %d, 错误: %d\n",
		response.TotalScanned, response.TotalFound, response.TotalErrors)

	// 显示状态码统计
	fmt.Println("状态码分布:")
	for statusCode, count := range response.StatusSummary {
		fmt.Printf("  %d: %d\n", statusCode, count)
	}

	// 显示前5个结果
	fmt.Println("扫描结果:")
	for i, result := range response.Results {
		if i >= 5 {
			fmt.Printf("... 还有 %d 个结果\n", len(response.Results)-5)
			break
		}
		fmt.Printf("  [%d] %s (长度: %d)\n",
			result.StatusCode, result.URL, result.ContentLength)
	}
}

// testCustomRealtimeScan 自定义配置实时扫描测试
func testCustomRealtimeScan() {
	start := time.Now()

	// 创建自定义配置
	options := &api.ScanOptions{
		URLs:           []string{"https://httpbin.org"},
		Wordlists:      []string{"https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt"},
		Threads:        5,   // 更少的线程数
		Timeout:        3.0, // 更短的超时
		Delay:          0.1, // 添加延迟
		ShowAllStatus:  true,
		RealTimeStatus: true,
		StatusFilter:   []int{200, 301, 302, 403, 404},
		UserAgent:      "Custom-API-Test/1.0",
	}

	fmt.Println("开始自定义配置实时扫描...")
	response, err := api.Scan(*options)
	if err != nil {
		log.Printf("自定义配置实时扫描失败: %v", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("自定义配置实时扫描完成，用时: %v\n", duration)
	fmt.Printf("总扫描数: %d, 发现: %d, 错误: %d\n",
		response.TotalScanned, response.TotalFound, response.TotalErrors)

	// 按状态码分组显示结果
	statusGroups := make(map[int][]api.ScanResult)
	for _, result := range response.Results {
		statusGroups[result.StatusCode] = append(statusGroups[result.StatusCode], result)
	}

	fmt.Println("按状态码分组的结果:")
	for statusCode, group := range statusGroups {
		fmt.Printf("  状态码 %d (%d 个):\n", statusCode, len(group))
		for i, result := range group {
			if i >= 3 { // 每个状态码只显示前3个
				fmt.Printf("    ... 还有 %d 个\n", len(group)-3)
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
