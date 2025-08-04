package view

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"dirsearch-go/internal/config"
	"dirsearch-go/internal/report"
)

// StatusDisplay 状态显示器
type StatusDisplay struct {
	config     *config.Config
	mu         sync.RWMutex
	startTime  time.Time
	totalPaths int
	scanned    int
	found      int
	errors     int
	status     map[int]int // 状态码统计
	lastUpdate time.Time
}

// NewStatusDisplay 创建新的状态显示器
func NewStatusDisplay(cfg *config.Config) *StatusDisplay {
	return &StatusDisplay{
		config:     cfg,
		status:     make(map[int]int),
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}
}

// SetTotalPaths 设置总路径数
func (sd *StatusDisplay) SetTotalPaths(total int) {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	sd.totalPaths = total
}

// UpdateProgress 更新进度
func (sd *StatusDisplay) UpdateProgress(result report.ScanResult) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	sd.scanned++

	if result.Error != nil {
		sd.errors++
	} else {
		sd.status[result.StatusCode]++
		if result.StatusCode >= 200 && result.StatusCode < 400 {
			sd.found++
		}
	}

	// 实时显示（如果启用）
	if sd.config.View.RealTimeStatus && time.Since(sd.lastUpdate) > time.Millisecond*500 {
		sd.displayProgress()
		sd.lastUpdate = time.Now()
	}
}

// DisplayFinalResults 显示最终结果
func (sd *StatusDisplay) DisplayFinalResults(results []report.ScanResult) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("扫描完成!")
	fmt.Println(strings.Repeat("=", 50))

	elapsed := time.Since(sd.startTime)
	fmt.Printf("扫描时间: %s\n", formatDuration(elapsed))
	fmt.Printf("总路径数: %d | 已扫描: %d | 发现: %d | 错误: %d\n",
		sd.totalPaths, sd.scanned, sd.found, sd.errors)

	// 显示状态码统计
	if len(sd.status) > 0 {
		fmt.Println("\n状态码分布:")
		for code, count := range sd.status {
			fmt.Printf("  %d: %d\n", code, count)
		}
	}

	fmt.Println(strings.Repeat("=", 50))
}

// displayProgress 显示进度
func (sd *StatusDisplay) displayProgress() {
	if sd.totalPaths == 0 {
		return
	}

	elapsed := time.Since(sd.startTime)
	progress := float64(sd.scanned) / float64(sd.totalPaths) * 100

	// 计算预估剩余时间
	var eta time.Duration
	if sd.scanned > 0 {
		rate := float64(sd.scanned) / elapsed.Seconds()
		remaining := float64(sd.totalPaths-sd.scanned) / rate
		eta = time.Duration(remaining) * time.Second
	}

	fmt.Printf("\r[%s] %.1f%% (%d/%d) | 发现: %d | 错误: %d | 用时: %s | 剩余: %s",
		getProgressBar(progress),
		progress,
		sd.scanned,
		sd.totalPaths,
		sd.found,
		sd.errors,
		formatDuration(elapsed),
		formatDuration(eta),
	)
}

// getProgressBar 获取进度条
func getProgressBar(progress float64) string {
	const width = 30
	filled := int(progress / 100 * width)

	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"

	return bar
}

// formatDuration 格式化时间
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	} else {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	}
}

// DisplayHeadlessSummary 显示无头模式摘要
func (sd *StatusDisplay) DisplayHeadlessSummary(results []report.ScanResult) {
	if !sd.config.View.Headless {
		return
	}

	fmt.Println("dirsearch-go 扫描摘要")
	fmt.Println("==================")
	fmt.Printf("目标数量: %d\n", len(results))
	fmt.Printf("扫描时间: %s\n", formatDuration(time.Since(sd.startTime)))

	// 统计状态码
	statusCount := make(map[int]int)
	for _, result := range results {
		statusCount[result.StatusCode]++
	}

	fmt.Println("状态码分布:")
	for code, count := range statusCount {
		fmt.Printf("  %d: %d\n", code, count)
	}
}
