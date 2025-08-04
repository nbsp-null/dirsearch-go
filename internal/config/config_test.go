package config

import (
	"os"
	"testing"
)

func TestParseStatusCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
		hasError bool
	}{
		{
			name:     "single code",
			input:    "200",
			expected: []int{200},
			hasError: false,
		},
		{
			name:     "multiple codes",
			input:    "200,301,404",
			expected: []int{200, 301, 404},
			hasError: false,
		},
		{
			name:     "range codes",
			input:    "200-299",
			expected: []int{200, 201, 202, 203, 204, 205, 206, 207, 208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 219, 220, 221, 222, 223, 224, 225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 235, 236, 237, 238, 239, 240, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257, 258, 259, 260, 261, 262, 263, 264, 265, 266, 267, 268, 269, 270, 271, 272, 273, 274, 275, 276, 277, 278, 279, 280, 281, 282, 283, 284, 285, 286, 287, 288, 289, 290, 291, 292, 293, 294, 295, 296, 297, 298, 299},
			hasError: false,
		},
		{
			name:     "mixed codes",
			input:    "200,301-303,404",
			expected: []int{200, 301, 302, 303, 404},
			hasError: false,
		},
		{
			name:     "invalid range",
			input:    "200-",
			expected: nil,
			hasError: true,
		},
		{
			name:     "invalid code",
			input:    "abc",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseStatusCodes(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d codes, got %d", len(tt.expected), len(result))
				return
			}

			for i, code := range result {
				if code != tt.expected[i] {
					t.Errorf("Expected code %d at position %d, got %d", tt.expected[i], i, code)
				}
			}
		})
	}
}

func TestInit(t *testing.T) {
	// 保存原始环境变量
	originalConfig := os.Getenv("DIRSEARCH_CONFIG")
	defer os.Setenv("DIRSEARCH_CONFIG", originalConfig)

	// 测试默认配置初始化
	err := Init()
	if err != nil {
		t.Errorf("Failed to initialize config: %v", err)
	}

	// 验证全局配置不为空
	if GlobalConfig == nil {
		t.Error("GlobalConfig should not be nil after initialization")
	}

	// 验证默认值
	if GlobalConfig.General.Threads != 25 {
		t.Errorf("Expected default threads to be 25, got %d", GlobalConfig.General.Threads)
	}

	if GlobalConfig.Connection.Timeout != 7.5 {
		t.Errorf("Expected default timeout to be 7.5, got %f", GlobalConfig.Connection.Timeout)
	}
}

func TestGetConfig(t *testing.T) {
	// 确保配置已初始化
	if GlobalConfig == nil {
		Init()
	}

	config := GetConfig()
	if config == nil {
		t.Error("GetConfig should not return nil")
	}
}
