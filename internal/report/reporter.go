package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"dirsearch-go/internal/config"
)

// ScanResult 扫描结果接口
type ScanResult struct {
	URL            string
	Path           string
	StatusCode     int
	Size           int64
	Title          string
	Redirect       string
	Error          error
	Timestamp      time.Time
	IsDirectory    bool
	RecursionLevel int
	Headers        http.Header
	Body           string
}

// Reporter 报告生成器
type Reporter struct {
	config *config.Config
}

// NewReporter 创建新的报告生成器
func NewReporter(cfg *config.Config) (*Reporter, error) {
	return &Reporter{
		config: cfg,
	}, nil
}

// SaveResults 保存扫描结果
func (r *Reporter) SaveResults(results []ScanResult, filename string) error {
	format := r.config.Output.ReportFormat
	if format == "" {
		format = "plain"
	}

	switch format {
	case "json":
		return r.saveJSON(results, filename)
	case "csv":
		return r.saveCSV(results, filename)
	case "html":
		return r.saveHTML(results, filename)
	case "plain":
		return r.savePlain(results, filename)
	case "simple":
		return r.saveSimple(results, filename)
	default:
		return fmt.Errorf("unsupported report format: %s", format)
	}
}

// saveJSON 保存JSON格式报告
func (r *Reporter) saveJSON(results []ScanResult, filename string) error {
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(results)
}

// saveCSV 保存CSV格式报告
func (r *Reporter) saveCSV(results []ScanResult, filename string) error {
	if !strings.HasSuffix(filename, ".csv") {
		filename += ".csv"
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	header := []string{"URL", "Path", "Status Code", "Size", "Title", "Redirect", "Error", "Timestamp"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// 写入数据
	for _, result := range results {
		row := []string{
			result.URL,
			result.Path,
			fmt.Sprintf("%d", result.StatusCode),
			fmt.Sprintf("%d", result.Size),
			result.Title,
			result.Redirect,
			"",
		}
		if result.Error != nil {
			row[6] = result.Error.Error()
		}
		row = append(row, result.Timestamp.Format(time.RFC3339))

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// saveHTML 保存HTML格式报告
func (r *Reporter) saveHTML(results []ScanResult, filename string) error {
	if !strings.HasSuffix(filename, ".html") {
		filename += ".html"
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// HTML模板
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <title>dirsearch-go Scan Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .status-200 { background-color: #d4edda; }
        .status-301, .status-302 { background-color: #fff3cd; }
        .status-403, .status-404 { background-color: #f8d7da; }
        .status-500 { background-color: #f5c6cb; }
    </style>
</head>
<body>
    <h1>dirsearch-go Scan Report</h1>
    <p>Generated: {{.Timestamp}}</p>
    <p>Total Results: {{len .Results}}</p>
    
    <table>
        <thead>
            <tr>
                <th>URL</th>
                <th>Path</th>
                <th>Status Code</th>
                <th>Size</th>
                <th>Title</th>
                <th>Redirect</th>
            </tr>
        </thead>
        <tbody>
            {{range .Results}}
            <tr class="status-{{.StatusCode}}">
                <td>{{.URL}}</td>
                <td>{{.Path}}</td>
                <td>{{.StatusCode}}</td>
                <td>{{.Size}}</td>
                <td>{{.Title}}</td>
                <td>{{.Redirect}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</body>
</html>`

	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		Results   []ScanResult
		Timestamp time.Time
	}{
		Results:   results,
		Timestamp: time.Now(),
	}

	return tmpl.Execute(file, data)
}

// savePlain 保存纯文本格式报告
func (r *Reporter) savePlain(results []ScanResult, filename string) error {
	if !strings.HasSuffix(filename, ".txt") {
		filename += ".txt"
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 写入报告头
	fmt.Fprintf(file, "dirsearch-go Scan Report\n")
	fmt.Fprintf(file, "Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "Total Results: %d\n\n", len(results))

	// 写入结果
	for _, result := range results {
		fmt.Fprintf(file, "[%d] %s%s\n", result.StatusCode, result.URL, result.Path)
		if result.Title != "" {
			fmt.Fprintf(file, "    Title: %s\n", result.Title)
		}
		if result.Redirect != "" {
			fmt.Fprintf(file, "    Redirect: %s\n", result.Redirect)
		}
		if result.Error != nil {
			fmt.Fprintf(file, "    Error: %s\n", result.Error.Error())
		}
		fmt.Fprintf(file, "\n")
	}

	return nil
}

// saveSimple 保存简单格式报告
func (r *Reporter) saveSimple(results []ScanResult, filename string) error {
	if !strings.HasSuffix(filename, ".txt") {
		filename += ".txt"
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 只输出状态码和路径
	for _, result := range results {
		fmt.Fprintf(file, "[%d] %s\n", result.StatusCode, result.Path)
	}

	return nil
}

// CreateReportDirectory 创建报告目录
func (r *Reporter) CreateReportDirectory() error {
	reportDir := r.config.Output.AutosaveReportFolder
	if reportDir == "" {
		reportDir = "reports"
	}

	return os.MkdirAll(reportDir, 0755)
}

// GenerateReportFilename 生成报告文件名
func (r *Reporter) GenerateReportFilename(format string) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("dirsearch_report_%s.%s", timestamp, format)
}
