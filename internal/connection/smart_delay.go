package connection

import (
	"fmt"
	"net"
	"time"

	"dirsearch-go/internal/config"
)

// SmartDelay 智能延迟管理器
type SmartDelay struct {
	config     *config.Config
	baseDelay  time.Duration
	multiplier float64
	pingDelay  time.Duration
}

// NewSmartDelay 创建智能延迟管理器
func NewSmartDelay(cfg *config.Config) *SmartDelay {
	return &SmartDelay{
		config:     cfg,
		baseDelay:  time.Duration(cfg.Connection.Delay * float64(time.Second)),
		multiplier: 10.0, // 基础延迟的10倍
	}
}

// MeasurePingDelay 测量ping延迟
func (sd *SmartDelay) MeasurePingDelay(host string) error {
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
	conn, err := net.DialTimeout("tcp", hostname+":80", 5*time.Second)
	if err != nil {
		// 尝试443端口
		conn, err = net.DialTimeout("tcp", hostname+":443", 5*time.Second)
		if err != nil {
			return fmt.Errorf("failed to ping %s: %w", hostname, err)
		}
	}
	defer conn.Close()

	sd.pingDelay = time.Since(start)
	return nil
}

// GetSmartDelay 获取智能延迟时间
func (sd *SmartDelay) GetSmartDelay() time.Duration {
	if sd.pingDelay > 0 {
		// 使用ping延迟的10倍作为连接延迟
		smartDelay := time.Duration(float64(sd.pingDelay) * sd.multiplier)

		// 设置最小和最大延迟限制
		minDelay := 100 * time.Millisecond
		maxDelay := 5 * time.Second

		if smartDelay < minDelay {
			smartDelay = minDelay
		} else if smartDelay > maxDelay {
			smartDelay = maxDelay
		}

		return smartDelay
	}

	// 如果没有ping延迟，使用配置的延迟
	return sd.baseDelay
}

// IsSlowResponse 判断是否为慢响应
func (sd *SmartDelay) IsSlowResponse(responseTime time.Duration) bool {
	if sd.pingDelay > 0 {
		// 如果响应时间超过ping延迟的20倍，认为是慢响应
		threshold := time.Duration(float64(sd.pingDelay) * 20.0)
		return responseTime > threshold
	}

	// 默认阈值：2秒
	return responseTime > 2*time.Second
}

// GetTimeout 获取超时时间
func (sd *SmartDelay) GetTimeout() time.Duration {
	if sd.pingDelay > 0 {
		// 使用ping延迟的30倍作为超时时间
		timeout := time.Duration(float64(sd.pingDelay) * 30.0)

		// 设置最小和最大超时限制
		minTimeout := 5 * time.Second
		maxTimeout := 30 * time.Second

		if timeout < minTimeout {
			timeout = minTimeout
		} else if timeout > maxTimeout {
			timeout = maxTimeout
		}

		return timeout
	}

	// 默认超时时间
	return time.Duration(sd.config.Connection.Timeout * float64(time.Second))
}

// GetPingDelay 获取ping延迟
func (sd *SmartDelay) GetPingDelay() time.Duration {
	return sd.pingDelay
}

// SetMultiplier 设置延迟倍数
func (sd *SmartDelay) SetMultiplier(multiplier float64) {
	sd.multiplier = multiplier
}

// GetMultiplier 获取延迟倍数
func (sd *SmartDelay) GetMultiplier() float64 {
	return sd.multiplier
}
