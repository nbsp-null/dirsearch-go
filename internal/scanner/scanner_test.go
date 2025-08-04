package scanner

import (
	"testing"

	"dirsearch-go/internal/config"
)

func TestSmartPathJoin(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{}
	scanner := &Scanner{config: cfg}

	tests := []struct {
		name     string
		base     string
		path     string
		expected string
	}{
		{
			name:     "基础URL不以分隔符结尾，路径不以分隔符开头",
			base:     "https://example.com",
			path:     "admin",
			expected: "https://example.com/admin",
		},
		{
			name:     "基础URL以分隔符结尾，路径不以分隔符开头",
			base:     "https://example.com/",
			path:     "admin",
			expected: "https://example.com/admin",
		},
		{
			name:     "基础URL不以分隔符结尾，路径以分隔符开头",
			base:     "https://example.com",
			path:     "/admin",
			expected: "https://example.com/admin",
		},
		{
			name:     "基础URL以分隔符结尾，路径以分隔符开头",
			base:     "https://example.com/",
			path:     "/admin",
			expected: "https://example.com/admin",
		},
		{
			name:     "基础URL以反斜杠结尾，路径以正斜杠开头",
			base:     "https://example.com\\",
			path:     "/admin",
			expected: "https://example.com/admin",
		},
		{
			name:     "基础URL以正斜杠结尾，路径以反斜杠开头",
			base:     "https://example.com/",
			path:     "\\admin",
			expected: "https://example.com/admin",
		},
		{
			name:     "空路径",
			base:     "https://example.com/",
			path:     "",
			expected: "https://example.com",
		},
		{
			name:     "复杂路径",
			base:     "https://example.com/api",
			path:     "/v1/users",
			expected: "https://example.com/api/v1/users",
		},
		{
			name:     "多层路径",
			base:     "https://example.com/",
			path:     "/admin/users/",
			expected: "https://example.com/admin/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.smartPathJoin(tt.base, tt.path)
			if result != tt.expected {
				t.Errorf("smartPathJoin(%q, %q) = %q, want %q", tt.base, tt.path, result, tt.expected)
			}
		})
	}
}

func TestBuildURL(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{}
	scanner := &Scanner{config: cfg}

	tests := []struct {
		name        string
		target      string
		path        string
		expectError bool
	}{
		{
			name:        "有效URL",
			target:      "https://example.com",
			path:        "admin",
			expectError: false,
		},
		{
			name:        "无效URL",
			target:      "invalid-url",
			path:        "admin",
			expectError: true,
		},
		{
			name:        "带路径分隔符",
			target:      "https://example.com/",
			path:        "/admin",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := scanner.buildURL(tt.target, tt.path)
			if tt.expectError && err == nil {
				t.Errorf("buildURL(%q, %q) expected error but got none", tt.target, tt.path)
			}
			if !tt.expectError && err != nil {
				t.Errorf("buildURL(%q, %q) unexpected error: %v", tt.target, tt.path, err)
			}
		})
	}
}
