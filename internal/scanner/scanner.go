package scanner

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"dirsearch-go/internal/config"
	"dirsearch-go/internal/connection"
	"dirsearch-go/internal/dictionary"
	"dirsearch-go/internal/report"
	"dirsearch-go/internal/view"
)

// ScanResult 扫描结果类型别名
type ScanResult = report.ScanResult

// Scanner 扫描器
type Scanner struct {
	config          *config.Config
	requester       *connection.Requester
	dictionary      *dictionary.Dictionary
	reporter        *report.Reporter
	domainChecker   *connection.DomainChecker
	headlessBrowser *connection.HeadlessBrowser
	statusDisplay   *view.StatusDisplay
	results         []ScanResult
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewScanner 创建新的扫描器
func NewScanner(cfg *config.Config) (*Scanner, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("NewScanner panic recovered: %v\nStack trace: %s", r, debug.Stack())
		}
	}()

	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建请求器
	requester, err := connection.NewRequester(cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create requester: %w", err)
	}

	// 创建字典
	dict, err := dictionary.NewDictionary(cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create dictionary: %w", err)
	}

	// 创建报告器
	reporter, err := report.NewReporter(cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create reporter: %w", err)
	}

	// 创建域名检查器
	domainChecker := connection.NewDomainChecker(cfg)

	// 创建状态显示器
	statusDisplay := view.NewStatusDisplay(cfg)

	// 创建无头浏览器（如果启用）
	var headlessBrowser *connection.HeadlessBrowser
	if cfg.View.Headless {
		var err error
		headlessBrowser, err = connection.NewHeadlessBrowser(cfg)
		if err != nil {
			log.Printf("Warning: Failed to create headless browser: %v", err)
		}
	}

	return &Scanner{
		config:          cfg,
		requester:       requester,
		dictionary:      dict,
		reporter:        reporter,
		domainChecker:   domainChecker,
		headlessBrowser: headlessBrowser,
		statusDisplay:   statusDisplay,
		results:         make([]ScanResult, 0),
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// Scan 执行扫描
func (s *Scanner) Scan(targets []string) ([]ScanResult, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Scan panic recovered: %v\nStack trace: %s", r, debug.Stack())
		}
	}()

	if targets == nil || len(targets) == 0 {
		return nil, fmt.Errorf("no targets specified")
	}

	// 域名存活检测
	fmt.Println("正在检测域名存活状态...")
	aliveTargets, deadTargets := s.domainChecker.CheckMultipleDomains(targets)

	// 显示不存活的域名
	if len(deadTargets) > 0 {
		fmt.Println("\n以下域名不存活:")
		for _, target := range deadTargets {
			fmt.Printf("  ❌ %s\n", target)
		}
		fmt.Println()
	}

	if len(aliveTargets) == 0 {
		return nil, fmt.Errorf("没有存活的域名可以扫描")
	}

	fmt.Printf("发现 %d 个存活域名，开始扫描...\n", len(aliveTargets))

	// 标准化URL，确保末尾有斜杠
	aliveTargets = s.normalizeTargets(aliveTargets)

	// 生成扫描路径
	paths, err := s.dictionary.GeneratePaths()
	if err != nil {
		return nil, fmt.Errorf("failed to generate paths: %w", err)
	}

	// 执行扫描
	results, err := s.executeScan(aliveTargets, paths, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to execute scan: %w", err)
	}

	// 显示最终结果
	s.statusDisplay.DisplayFinalResults(results)

	// 如果是无头模式，显示摘要
	s.statusDisplay.DisplayHeadlessSummary(results)

	return results, nil
}

// executeScan 执行扫描（支持递归）
func (s *Scanner) executeScan(targets []string, paths []string, recursionLevel int) ([]ScanResult, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("executeScan panic recovered: %v\nStack trace: %s", r, debug.Stack())
		}
	}()

	if targets == nil || len(targets) == 0 {
		return []ScanResult{}, nil
	}

	if paths == nil || len(paths) == 0 {
		return []ScanResult{}, nil
	}

	// 设置状态显示器的总路径数
	totalPaths := len(targets) * len(paths)
	s.statusDisplay.SetTotalPaths(totalPaths)

	// 创建工作池
	workerCount := s.config.General.Threads
	if workerCount <= 0 {
		workerCount = 25 // 默认线程数
	}

	// 创建任务通道
	taskChan := make(chan ScanTask, workerCount*2)
	resultChan := make(chan ScanResult, workerCount*2)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Worker %d panic recovered: %v", workerID, r)
				}
			}()
			defer wg.Done()
			s.worker(&wg, taskChan, resultChan)
		}(i)
	}

	// 发送任务
	go func() {
		defer close(taskChan)
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Task sender panic recovered: %v", r)
			}
		}()

		for _, target := range targets {
			for _, path := range paths {
				select {
				case taskChan <- ScanTask{Target: target, Path: path}:
				case <-s.ctx.Done():
					return
				}
			}
		}
	}()

	// 收集结果
	var results []ScanResult
	go func() {
		defer close(resultChan)
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Result collector panic recovered: %v", r)
			}
		}()

		for result := range resultChan {
			result.RecursionLevel = recursionLevel
			results = append(results, result)
			s.statusDisplay.UpdateProgress(result)
		}
	}()

	// 等待所有工作协程完成
	wg.Wait()

	// 如果启用递归扫描，对目录进行递归
	if s.config.View.RecursiveScan && recursionLevel < 3 { // 限制递归深度为3
		recursiveResults := s.performRecursiveScan(results, recursionLevel+1)
		results = append(results, recursiveResults...)
	}

	return results, nil
}

// performRecursiveScan 执行递归扫描
func (s *Scanner) performRecursiveScan(results []ScanResult, recursionLevel int) []ScanResult {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("performRecursiveScan panic recovered: %v", r)
		}
	}()

	var recursiveResults []ScanResult
	var directories []string

	// 识别目录（200或403状态码）
	for _, result := range results {
		if (result.StatusCode == 200 || result.StatusCode == 403) && s.isDirectory(result) {
			result.IsDirectory = true
			directories = append(directories, result.URL)
		}
	}

	if len(directories) == 0 {
		return recursiveResults
	}

	fmt.Printf("发现 %d 个目录，开始递归扫描...\n", len(directories))

	// 为每个目录生成子路径
	for _, directory := range directories {
		subPaths, err := s.dictionary.GeneratePaths() // 使用相同的字典
		if err != nil {
			log.Printf("Failed to generate paths for directory %s: %v", directory, err)
			continue
		}
		subResults, err := s.executeScan([]string{directory}, subPaths, recursionLevel)
		if err != nil {
			log.Printf("Failed to scan directory %s: %v", directory, err)
			continue // 忽略递归扫描错误
		}
		recursiveResults = append(recursiveResults, subResults...)
	}

	return recursiveResults
}

// isDirectory 判断是否为目录
func (s *Scanner) isDirectory(result ScanResult) bool {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("isDirectory panic recovered: %v", r)
		}
	}()

	// 检查URL是否以斜杠结尾
	if strings.HasSuffix(result.URL, "/") {
		return true
	}

	// 检查响应头中的Content-Type
	if result.Headers != nil {
		contentType := result.Headers.Get("Content-Type")
		if strings.Contains(contentType, "text/html") {
			// 检查HTML内容是否包含目录列表特征
			if strings.Contains(result.Body, "<title>Index of") ||
				strings.Contains(result.Body, "Directory listing for") ||
				strings.Contains(result.Body, "Parent Directory") {
				return true
			}
		}
	}

	return false
}

// normalizeTargets 标准化目标URL，确保末尾有斜杠
func (s *Scanner) normalizeTargets(targets []string) []string {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("normalizeTargets panic recovered: %v", r)
		}
	}()

	if targets == nil {
		return []string{}
	}

	normalized := make([]string, len(targets))
	for i, target := range targets {
		// 移除末尾的斜杠
		target = strings.TrimSuffix(target, "/")
		target = strings.TrimSuffix(target, "\\")
		// 添加斜杠
		normalized[i] = target + "/"
	}
	return normalized
}

// ScanTask 扫描任务
type ScanTask struct {
	Target string
	Path   string
}

// worker 工作协程
func (s *Scanner) worker(wg *sync.WaitGroup, taskChan <-chan ScanTask, resultChan chan<- ScanResult) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker panic recovered: %v", r)
		}
	}()

	for task := range taskChan {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// 使用安全的扫描方式
		result := s.scanPath(task.Target, task.Path)

		// 应用智能延迟
		if s.config.Connection.Delay > 0 {
			// 从URL中提取主机名
			if parsedURL, err := url.Parse(result.URL); err == nil {
				smartDelay := s.requester.HostManager.GetSmartDelay(parsedURL.Host)
				time.Sleep(smartDelay)
			}
		}

		select {
		case resultChan <- result:
		case <-s.ctx.Done():
			return
		}
	}
}

// scanPath 扫描单个路径
func (s *Scanner) scanPath(target, path string) ScanResult {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("scanPath panic recovered: %v", r)
		}
	}()

	result := ScanResult{
		URL:       target,
		Path:      path,
		Timestamp: time.Now(),
	}

	// 构建完整URL
	fullURL, err := s.buildURL(target, path)
	if err != nil {
		result.Error = fmt.Errorf("failed to build URL: %w", err)
		return result
	}

	// 根据模式选择扫描方法
	if s.config.View.Headless && s.headlessBrowser != nil {
		// 使用headless浏览器扫描
		headlessResult := s.headlessBrowser.ScanURL(fullURL)
		if headlessResult.Error != nil {
			result.Error = headlessResult.Error
		} else {
			result.StatusCode = headlessResult.StatusCode
			result.Size = headlessResult.ContentLength
			result.Title = headlessResult.Title
			result.Redirect = strings.Join(headlessResult.Redirects, " -> ")
		}
	} else {
		// 使用普通HTTP请求
		resp, err := s.requester.Request(fullURL)
		if err != nil {
			result.Error = fmt.Errorf("request failed: %w", err)
			return result
		}

		// 处理响应
		result.StatusCode = resp.StatusCode
		result.Size = resp.ContentLength
		result.Title = s.extractTitle(resp.Body)
		result.Redirect = resp.Redirect
		result.Headers = resp.Headers
		result.Body = resp.Body
	}

	return result
}

// buildURL 构建完整URL
func (s *Scanner) buildURL(target, path string) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("buildURL panic recovered: %v", r)
		}
	}()

	// 智能添加路径分隔符
	fullURL := s.smartPathJoin(target, path)

	// 验证URL格式
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// 检查URL是否有有效的scheme和host
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", fmt.Errorf("invalid URL: missing scheme or host")
	}

	return fullURL, nil
}

// smartPathJoin 智能路径拼接
func (s *Scanner) smartPathJoin(base, path string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("smartPathJoin panic recovered: %v", r)
		}
	}()

	// 标准化基础URL，移除末尾的斜杠
	base = strings.TrimSuffix(base, "/")
	base = strings.TrimSuffix(base, "\\")

	// 如果wordlist路径为空，直接返回基础URL
	if path == "" {
		return base + "/"
	}

	// 如果wordlist路径已经包含斜杠，直接拼接
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return base + path
	}

	// 如果wordlist路径以斜杠结尾，添加斜杠
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\") {
		return base + "/" + path
	}

	// 如果wordlist路径包含斜杠，添加斜杠
	if strings.Contains(path, "/") || strings.Contains(path, "\\") {
		return base + "/" + path
	}

	// 默认情况，添加斜杠
	return base + "/" + path
}

// extractTitle 提取页面标题
func (s *Scanner) extractTitle(body string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("extractTitle panic recovered: %v", r)
		}
	}()

	if body == "" {
		return ""
	}

	// 简单的标题提取逻辑
	titleStart := strings.Index(body, "<title>")
	if titleStart == -1 {
		return ""
	}

	titleStart += 7 // "<title>" 的长度
	titleEnd := strings.Index(body[titleStart:], "</title>")
	if titleEnd == -1 {
		return ""
	}

	title := body[titleStart : titleStart+titleEnd]
	return strings.TrimSpace(title)
}

// addResult 添加结果
func (s *Scanner) addResult(result ScanResult) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("addResult panic recovered: %v", r)
		}
	}()

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否应该包含此结果
	if s.shouldIncludeResult(result) {
		s.results = append(s.results, result)
	}
}

// shouldIncludeResult 检查是否应该包含结果
func (s *Scanner) shouldIncludeResult(result ScanResult) bool {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("shouldIncludeResult panic recovered: %v", r)
		}
	}()

	// 检查状态码过滤
	if len(s.config.General.IncludeStatus) > 0 {
		found := false
		for _, statusStr := range s.config.General.IncludeStatus {
			statusCodes, err := config.ParseStatusCodes(statusStr)
			if err != nil {
				continue
			}
			for _, status := range statusCodes {
				if result.StatusCode == status {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查排除状态码
	for _, statusStr := range s.config.General.ExcludeStatus {
		statusCodes, err := config.ParseStatusCodes(statusStr)
		if err != nil {
			continue
		}
		for _, status := range statusCodes {
			if result.StatusCode == status {
				return false
			}
		}
	}

	return true
}

// GetResults 获取结果
func (s *Scanner) GetResults() []ScanResult {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetResults panic recovered: %v", r)
		}
	}()

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 创建副本以避免并发问题
	results := make([]ScanResult, len(s.results))
	copy(results, s.results)
	return results
}

// Stop 停止扫描器
func (s *Scanner) Stop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Stop panic recovered: %v", r)
		}
	}()

	if s.cancel != nil {
		s.cancel()
	}

	// 清理资源
	if s.headlessBrowser != nil {
		s.headlessBrowser.Close()
	}
}

// SaveResults 保存结果
func (s *Scanner) SaveResults(filename string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("SaveResults panic recovered: %v", r)
		}
	}()

	results := s.GetResults()
	return s.reporter.SaveResults(results, filename)
}
