# DirSearch-Go API 文档

## 概述
DirSearch-Go提供了简洁的API接口，允许外部程序集成目录扫描功能。

## 核心API函数

### 1. Scan(options ScanOptions) (*ScanResponse, error)
主要的扫描函数，支持完整的扫描配置。

**参数:**
- `options`: 扫描选项结构体

**返回:**
- `*ScanResponse`: 扫描响应
- `error`: 错误信息

**示例:**
```go
options := api.ScanOptions{
    URLs:      []string{"https://example.com"},
    Wordlists: []string{"wordlists/common.txt"},
    Threads:   25,
    Timeout:   7.5,
}
response, err := api.Scan(options)
```

### 2. QuickScan(urls []string, wordlists []string, statusCodes []int) ([]ScanResult, error)
快速扫描函数，使用默认配置。

**参数:**
- `urls`: 目标URL列表
- `wordlists`: 字典文件列表
- `statusCodes`: 要包含的状态码列表

**返回:**
- `[]ScanResult`: 扫描结果列表
- `error`: 错误信息

**示例:**
```go
results, err := api.QuickScan(
    []string{"https://example.com"},
    []string{"wordlists/common.txt"},
    []int{200, 403},
)
```

### 3. ScanSingleURL(url string, wordlists []string, statusCodes []int) ([]ScanResult, error)
扫描单个URL的便捷函数。

**参数:**
- `url`: 目标URL
- `wordlists`: 字典文件列表
- `statusCodes`: 要包含的状态码列表

**返回:**
- `[]ScanResult`: 扫描结果列表
- `error`: 错误信息

**示例:**
```go
results, err := api.ScanSingleURL(
    "https://example.com",
    []string{"wordlists/common.txt"},
    []int{200, 403},
)
```

### 4. ScanSingleURLWithWordlist(url string, wordlistURL string, statusCodes []int) ([]ScanResult, error)
使用URL作为wordlist源扫描单个URL。

**参数:**
- `url`: 目标URL
- `wordlistURL`: wordlist的URL地址
- `statusCodes`: 要包含的状态码列表（可选）

**返回:**
- `[]ScanResult`: 扫描结果列表
- `error`: 错误信息

**示例:**
```go
results, err := api.ScanSingleURLWithWordlist(
    "https://example.com",
    "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt",
    []int{200, 403},
)
```

### 5. ScanSingleURLWithWordlistAdvanced(url string, wordlistURL string, options *ScanOptions) (*ScanResponse, error)
使用URL作为wordlist源扫描单个URL（高级选项）。

**参数:**
- `url`: 目标URL
- `wordlistURL`: wordlist的URL地址
- `options`: 高级扫描选项

**返回:**
- `*ScanResponse`: 完整的扫描响应
- `error`: 错误信息

**示例:**
```go
options := &api.ScanOptions{
    Threads:       10,
    Timeout:       5.0,
    ShowAllStatus: true,
    RecursiveScan: true,
}
response, err := api.ScanSingleURLWithWordlistAdvanced(
    "https://example.com",
    "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt",
    options,
)
```

## 数据结构

### ScanOptions
扫描选项结构体：
```go
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
```

### ScanResult
扫描结果结构体：
```go
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
```

### ScanResponse
扫描响应结构体：
```go
type ScanResponse struct {
    Results       []ScanResult `json:"results"`        // 扫描结果
    TotalScanned  int          `json:"total_scanned"`  // 总扫描数
    TotalFound    int          `json:"total_found"`    // 总发现数
    TotalErrors   int          `json:"total_errors"`   // 总错误数
    ScanTime      float64      `json:"scan_time"`      // 扫描时间(秒)
    StatusSummary map[int]int  `json:"status_summary"` // 状态码统计
}
```

## 使用示例

### 基本扫描
```go
package main

import (
    "fmt"
    "log"
    "dirsearch-go/internal/api"
)

func main() {
    // 基本扫描
    results, err := api.ScanSingleURL(
        "https://example.com",
        []string{"wordlists/common.txt"},
        []int{200, 403},
    )
    if err != nil {
        log.Fatal(err)
    }

    for _, result := range results {
        fmt.Printf("[%d] %s\n", result.StatusCode, result.URL)
    }
}
```

### 使用URL wordlist
```go
package main

import (
    "fmt"
    "log"
    "dirsearch-go/internal/api"
)

func main() {
    // 使用URL作为wordlist源
    results, err := api.ScanSingleURLWithWordlist(
        "https://example.com",
        "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt",
        []int{200, 403},
    )
    if err != nil {
        log.Fatal(err)
    }

    for _, result := range results {
        fmt.Printf("[%d] %s\n", result.StatusCode, result.URL)
    }
}
```

### 高级扫描
```go
package main

import (
    "fmt"
    "log"
    "dirsearch-go/internal/api"
)

func main() {
    // 高级扫描选项
    options := api.ScanOptions{
        URLs:          []string{"https://example.com"},
        Wordlists:     []string{"wordlists/common.txt"},
        Threads:       10,
        Timeout:       5.0,
        ShowAllStatus: true,
        RecursiveScan: true,
        UserAgent:     "Custom User Agent",
    }

    response, err := api.Scan(options)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("扫描完成: 发现 %d 个结果\n", response.TotalFound)
    for _, result := range response.Results {
        fmt.Printf("[%d] %s\n", result.StatusCode, result.URL)
    }
}
```

## 错误处理
所有API函数都包含完善的错误处理和panic恢复机制。建议在使用时检查返回的错误：

```go
results, err := api.ScanSingleURLWithWordlist(
    "https://example.com",
    "https://example.com/wordlist.txt",
    []int{200, 403},
)
if err != nil {
    log.Printf("扫描失败: %v", err)
    return
}
```

## 注意事项
1. 所有API函数都是线程安全的
2. 支持URL作为wordlist源，自动检测URL格式
3. 包含完善的异常处理，不会因为单个错误导致程序崩溃
4. 支持自定义状态码过滤
5. 提供实时状态显示选项 