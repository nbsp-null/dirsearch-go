package connection

import (
	"fmt"
	"log"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"dirsearch-go/internal/config"
)

// HostInfo 主机信息
type HostInfo struct {
	PingDelay  time.Duration
	LastPing   time.Time
	IsAlive    bool
	SmartDelay *SmartDelay
}

// HostManager 主机管理器
type HostManager struct {
	hosts  map[string]*HostInfo
	mu     sync.RWMutex
	config *config.Config
}

// NewHostManager 创建主机管理器
func NewHostManager(cfg *config.Config) *HostManager {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("NewHostManager panic recovered: %v\nStack trace: %s", r, debug.Stack())
		}
	}()

	if cfg == nil {
		// 返回默认配置而不是panic
		cfg = &config.Config{
			Connection: config.ConnectionConfig{
				Timeout: 7.5,
			},
		}
	}

	return &HostManager{
		hosts:  make(map[string]*HostInfo),
		config: cfg,
	}
}

// GetOrCreateHostInfo 获取或创建主机信息
func (hm *HostManager) GetOrCreateHostInfo(host string) *HostInfo {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetOrCreateHostInfo panic recovered: %v", r)
		}
	}()

	if host == "" {
		return &HostInfo{
			SmartDelay: NewSmartDelay(hm.config),
			IsAlive:    false,
		}
	}

	hm.mu.Lock()
	defer hm.mu.Unlock()

	// 检查是否已存在
	if info, exists := hm.hosts[host]; exists {
		return info
	}

	// 创建新的主机信息
	info := &HostInfo{
		SmartDelay: NewSmartDelay(hm.config),
		IsAlive:    false,
	}

	// 进行ping验证
	if err := hm.pingHost(host, info); err != nil {
		// ping失败，使用默认配置
		log.Printf("Ping failed for host %s: %v", host, err)
		info.PingDelay = 0
		info.IsAlive = false
	} else {
		info.IsAlive = true
	}

	// 缓存结果
	hm.hosts[host] = info
	return info
}

// pingHost 对主机进行ping验证
func (hm *HostManager) pingHost(host string, info *HostInfo) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("pingHost panic recovered: %v", r)
		}
	}()

	// 解析主机名
	hostname := host
	if hostname == "" {
		return fmt.Errorf("empty host")
	}

	// 移除端口号（如果有）
	if host, _, err := net.SplitHostPort(hostname); err == nil {
		hostname = host
	}

	// 测量ping延迟
	start := time.Now()

	// 尝试80端口
	conn, err := net.DialTimeout("tcp", hostname+":80", 5*time.Second)
	if err != nil {
		// 尝试443端口
		conn, err = net.DialTimeout("tcp", hostname+":443", 5*time.Second)
		if err != nil {
			return fmt.Errorf("failed to ping %s: %w", hostname, err)
		}
	}
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	info.PingDelay = time.Since(start)
	info.LastPing = time.Now()

	// 更新SmartDelay的ping延迟
	info.SmartDelay.pingDelay = info.PingDelay

	return nil
}

// GetSmartDelay 获取主机的智能延迟
func (hm *HostManager) GetSmartDelay(host string) time.Duration {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetSmartDelay panic recovered: %v", r)
		}
	}()

	info := hm.GetOrCreateHostInfo(host)
	if info == nil || info.SmartDelay == nil {
		return time.Duration(hm.config.Connection.Delay) * time.Second
	}
	return info.SmartDelay.GetSmartDelay()
}

// GetTimeout 获取主机的超时时间
func (hm *HostManager) GetTimeout(host string) time.Duration {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetTimeout panic recovered: %v", r)
		}
	}()

	info := hm.GetOrCreateHostInfo(host)
	if info == nil || info.SmartDelay == nil {
		return time.Duration(hm.config.Connection.Timeout) * time.Second
	}
	return info.SmartDelay.GetTimeout()
}

// IsSlowResponse 判断是否为慢响应
func (hm *HostManager) IsSlowResponse(host string, responseTime time.Duration) bool {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("IsSlowResponse panic recovered: %v", r)
		}
	}()

	info := hm.GetOrCreateHostInfo(host)
	if info == nil || info.SmartDelay == nil {
		return false
	}
	return info.SmartDelay.IsSlowResponse(responseTime)
}

// GetHostStats 获取主机统计信息
func (hm *HostManager) GetHostStats() map[string]*HostInfo {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetHostStats panic recovered: %v", r)
		}
	}()

	hm.mu.RLock()
	defer hm.mu.RUnlock()

	stats := make(map[string]*HostInfo)
	for host, info := range hm.hosts {
		stats[host] = info
	}
	return stats
}

// ClearCache 清除缓存
func (hm *HostManager) ClearCache() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ClearCache panic recovered: %v", r)
		}
	}()

	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.hosts = make(map[string]*HostInfo)
}
