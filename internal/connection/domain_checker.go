package connection

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"dirsearch-go/internal/config"
)

// DomainChecker 域名检测器
type DomainChecker struct {
	config *config.Config
	client *http.Client
}

// NewDomainChecker 创建新的域名检测器
func NewDomainChecker(cfg *config.Config) *DomainChecker {
	client := &http.Client{
		Timeout: time.Duration(cfg.Connection.DomainCheckTimeout * float64(time.Second)),
	}

	return &DomainChecker{
		config: cfg,
		client: client,
	}
}

// CheckDomain 检测域名是否存活
func (dc *DomainChecker) CheckDomain(targetURL string) (bool, error) {
	// 解析URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return false, fmt.Errorf("invalid URL: %w", err)
	}

	// 构建检测URL（使用根路径）
	checkURL := fmt.Sprintf("%s://%s/", parsedURL.Scheme, parsedURL.Host)
	if parsedURL.Scheme == "" {
		checkURL = fmt.Sprintf("http://%s/", parsedURL.Host)
	}

	// 重试检测
	for attempt := 1; attempt <= dc.config.Connection.DomainCheckRetries; attempt++ {
		if dc.isDomainAlive(checkURL) {
			return true, nil
		}

		if attempt < dc.config.Connection.DomainCheckRetries {
			// 等待一段时间后重试
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	return false, fmt.Errorf("domain not alive after %d attempts", dc.config.Connection.DomainCheckRetries)
}

// isDomainAlive 检测单个域名是否存活
func (dc *DomainChecker) isDomainAlive(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dc.config.Connection.DomainCheckTimeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false
	}

	// 设置基本请求头
	req.Header.Set("User-Agent", "dirsearch-go/1.0")
	req.Header.Set("Accept", "*/*")

	resp, err := dc.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 检查状态码，2xx和3xx都认为是存活的
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

// CheckMultipleDomains 批量检测域名
func (dc *DomainChecker) CheckMultipleDomains(targets []string) ([]string, []string) {
	var aliveTargets []string
	var deadTargets []string

	for _, target := range targets {
		if alive, _ := dc.CheckDomain(target); alive {
			aliveTargets = append(aliveTargets, target)
		} else {
			deadTargets = append(deadTargets, target)
		}
	}

	return aliveTargets, deadTargets
}
