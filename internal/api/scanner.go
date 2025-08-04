package api

import (
	"fmt"
	"log"
	"runtime/debug"

	"dirsearch-go/internal/config"
	"dirsearch-go/internal/report"
	"dirsearch-go/internal/scanner"
)

// ScanOptions 扫描选项
type ScanOptions struct {
	// 基本设置
	URLs      []string `json:"urls"`      // 目标URL列表
	Wordlists []string `json:"wordlists"` // 字典文件列表
	Threads   int      `json:"threads"`   // 线程数
	Delay     float64  `json:"delay"`     // 请求延迟

	// 输出控制
	ShowAllStatus bool  `json:"show_all_status"` // 是否显示所有状态码
	StatusFilter  []int `json:"status_filter"`   // 指定状态码过滤
	RecursiveScan bool  `json:"recursive_scan"`  // 是否启用递归扫描

	// 请求设置
	UserAgent string   `json:"user_agent"` // 用户代理
	Headers   []string `json:"headers"`    // 请求头
	Proxy     string   `json:"proxy"`      // 代理设置
	Timeout   float64  `json:"timeout"`    // 超时时间

	// 高级设置
	RealTimeStatus bool `json:"real_time_status"` // 实时状态显示
	Headless       bool `json:"headless"`         // 无头模式
}

// ScanResult 扫描结果
type ScanResult struct {
	URL            string            `json:"url"`             // 完整URL
	Path           string            `json:"path"`            // 扫描路径
	StatusCode     int               `json:"status_code"`     // HTTP状态码
	ContentLength  int64             `json:"content_length"`  // 内容长度
	Title          string            `json:"title"`           // 页面标题
	Redirect       string            `json:"redirect"`        // 重定向URL
	Headers        map[string]string `json:"headers"`         // 响应头
	Body           string            `json:"body"`            // 响应体
	IsDirectory    bool              `json:"is_directory"`    // 是否为目录
	RecursionLevel int               `json:"recursion_level"` // 递归层级
	Error          string            `json:"error,omitempty"` // 错误信息
}

// ScanResponse 扫描响应
type ScanResponse struct {
	Results       []ScanResult `json:"results"`        // 扫描结果
	TotalScanned  int          `json:"total_scanned"`  // 总扫描数
	TotalFound    int          `json:"total_found"`    // 总发现数
	TotalErrors   int          `json:"total_errors"`   // 总错误数
	ScanTime      float64      `json:"scan_time"`      // 扫描时间(秒)
	StatusSummary map[int]int  `json:"status_summary"` // 状态码统计
}

// Scan 执行扫描
// 参数:
//   - options: 扫描选项
//
// 返回:
//   - ScanResponse: 扫描响应
//   - error: 错误信息
func Scan(options ScanOptions) (*ScanResponse, error) {
	// 使用defer和recover捕获panic
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Scan panic recovered: %v\nStack trace: %s", r, debug.Stack())
		}
	}()

	// 验证输入参数
	if err := validateOptions(&options); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// 创建配置
	cfg := createConfig(&options)

	// 创建扫描器
	scanner, err := scanner.NewScanner(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create scanner: %w", err)
	}

	// 使用defer确保扫描器资源被正确释放
	defer func() {
		if scanner != nil {
			scanner.Stop()
		}
	}()

	// 执行扫描
	results, err := scanner.Scan(options.URLs)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// 转换结果格式（添加异常处理）
	apiResults, err := convertResults(results)
	if err != nil {
		return nil, fmt.Errorf("failed to convert results: %w", err)
	}

	// 应用状态码过滤
	if len(options.StatusFilter) > 0 {
		apiResults = filterByStatus(apiResults, options.StatusFilter)
	}

	// 构建响应
	response := buildResponse(apiResults, results)

	return response, nil
}

// validateOptions 验证扫描选项
func validateOptions(options *ScanOptions) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("validateOptions panic recovered: %v", r)
		}
	}()

	if options == nil {
		return fmt.Errorf("options cannot be nil")
	}

	if len(options.URLs) == 0 {
		return fmt.Errorf("no URLs specified")
	}

	if len(options.Wordlists) == 0 {
		return fmt.Errorf("no wordlists specified")
	}

	if options.Threads <= 0 {
		options.Threads = 25 // 默认线程数
	}

	if options.Timeout <= 0 {
		options.Timeout = 7.5 // 默认超时时间
	}

	return nil
}

// createConfig 根据选项创建配置
func createConfig(options *ScanOptions) *config.Config {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("createConfig panic recovered: %v", r)
		}
	}()

	if options == nil {
		// 返回默认配置而不是panic
		return &config.Config{
			General: config.GeneralConfig{
				Threads: 25,
			},
			Connection: config.ConnectionConfig{
				Timeout: 7.5,
			},
			Request: config.RequestConfig{
				HTTPMethod: "GET",
			},
		}
	}

	cfg := &config.Config{
		General: config.GeneralConfig{
			Threads: options.Threads,
		},
		Dictionary: config.DictionaryConfig{
			Wordlists: options.Wordlists,
		},
		Connection: config.ConnectionConfig{
			Delay:   options.Delay,
			Timeout: options.Timeout,
		},
		Request: config.RequestConfig{
			HTTPMethod: "GET",
			UserAgent:  options.UserAgent,
			Headers:    options.Headers,
		},
		View: config.ViewConfig{
			ShowAllStatus:  options.ShowAllStatus,
			RecursiveScan:  options.RecursiveScan,
			RealTimeStatus: options.RealTimeStatus,
			Headless:       options.Headless,
		},
	}

	// 设置代理
	if options.Proxy != "" {
		cfg.Connection.Proxy = options.Proxy
	}

	return cfg
}

// convertResults 转换结果格式
func convertResults(results []report.ScanResult) ([]ScanResult, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("convertResults panic recovered: %v", r)
		}
	}()

	if results == nil {
		return []ScanResult{}, nil
	}

	apiResults := make([]ScanResult, 0, len(results))

	for i, result := range results {
		// 使用安全的转换方式
		apiResult, err := convertSingleResult(result)
		if err != nil {
			log.Printf("Failed to convert result %d: %v", i, err)
			continue // 跳过有问题的结果，而不是整个失败
		}
		apiResults = append(apiResults, apiResult)
	}

	return apiResults, nil
}

// convertSingleResult 转换单个结果
func convertSingleResult(result report.ScanResult) (ScanResult, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("convertSingleResult panic recovered: %v", r)
		}
	}()

	// 转换响应头
	headers := make(map[string]string)
	if result.Headers != nil {
		for key, values := range result.Headers {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
	}

	apiResult := ScanResult{
		URL:            result.URL,
		Path:           result.Path,
		StatusCode:     result.StatusCode,
		ContentLength:  result.Size,
		Title:          result.Title,
		Redirect:       result.Redirect,
		Headers:        headers,
		Body:           result.Body,
		IsDirectory:    result.IsDirectory,
		RecursionLevel: result.RecursionLevel,
		Error:          "",
	}

	// 处理错误
	if result.Error != nil {
		apiResult.Error = result.Error.Error()
	}

	return apiResult, nil
}

// filterByStatus 根据状态码过滤结果
func filterByStatus(results []ScanResult, statusCodes []int) []ScanResult {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("filterByStatus panic recovered: %v", r)
		}
	}()

	if results == nil || statusCodes == nil {
		return results
	}

	var filtered []ScanResult

	for _, result := range results {
		for _, code := range statusCodes {
			if result.StatusCode == code {
				filtered = append(filtered, result)
				break
			}
		}
	}

	return filtered
}

// buildResponse 构建扫描响应
func buildResponse(apiResults []ScanResult, originalResults []report.ScanResult) *ScanResponse {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("buildResponse panic recovered: %v", r)
		}
	}()

	// 统计状态码
	statusSummary := make(map[int]int)
	for _, result := range apiResults {
		statusSummary[result.StatusCode]++
	}

	// 统计错误数
	errorCount := 0
	for _, result := range apiResults {
		if result.Error != "" {
			errorCount++
		}
	}

	return &ScanResponse{
		Results:       apiResults,
		TotalScanned:  len(originalResults),
		TotalFound:    len(apiResults),
		TotalErrors:   errorCount,
		ScanTime:      0, // 需要从扫描器获取实际时间
		StatusSummary: statusSummary,
	}
}

// QuickScan 快速扫描函数
// 参数:
//   - urls: 目标URL列表
//   - wordlists: 字典文件列表
//   - statusCodes: 要显示的状态码列表(为空则显示所有)
//
// 返回:
//   - []ScanResult: 扫描结果
//   - error: 错误信息
func QuickScan(urls []string, wordlists []string, statusCodes []int) ([]ScanResult, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("QuickScan panic recovered: %v", r)
		}
	}()

	options := ScanOptions{
		URLs:          urls,
		Wordlists:     wordlists,
		Threads:       25,
		Delay:         0,
		ShowAllStatus: len(statusCodes) == 0,
		StatusFilter:  statusCodes,
		Timeout:       7.5,
	}

	response, err := Scan(options)
	if err != nil {
		return nil, err
	}

	return response.Results, nil
}

// ScanSingleURL 扫描单个URL
// 参数:
//   - url: 目标URL
//   - wordlists: 字典文件列表
//   - statusCodes: 要显示的状态码列表(为空则显示所有)
//
// 返回:
//   - []ScanResult: 扫描结果
//   - error: 错误信息
func ScanSingleURL(url string, wordlists []string, statusCodes []int) ([]ScanResult, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ScanSingleURL panic recovered: %v", r)
		}
	}()

	return QuickScan([]string{url}, wordlists, statusCodes)
}
