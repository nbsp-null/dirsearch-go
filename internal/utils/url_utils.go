package utils

import (
	"net/url"
	"strings"
)

// IsURL 检查字符串是否为有效的URL
func IsURL(str string) bool {
	// 检查是否包含协议
	if strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://") {
		_, err := url.ParseRequestURI(str)
		return err == nil
	}
	return false
}

// NormalizeURL 标准化URL，确保正确处理斜杠
func NormalizeURL(url string) string {
	// 移除末尾的斜杠
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, "\\")
	return url
}

// BuildPath 构建完整路径，智能处理斜杠
func BuildPath(baseURL, wordlistPath string) string {
	// 标准化基础URL
	baseURL = NormalizeURL(baseURL)

	// 标准化wordlist路径
	wordlistPath = strings.TrimPrefix(wordlistPath, "/")
	wordlistPath = strings.TrimPrefix(wordlistPath, "\\")

	// 如果wordlist路径为空，直接返回基础URL
	if wordlistPath == "" {
		return baseURL + "/"
	}

	// 检查wordlist路径是否以斜杠开头
	if strings.HasPrefix(wordlistPath, "/") || strings.HasPrefix(wordlistPath, "\\") {
		// 如果wordlist路径以斜杠开头，直接拼接
		return baseURL + wordlistPath
	}

	// 检查wordlist路径是否以斜杠结尾
	if strings.HasSuffix(wordlistPath, "/") || strings.HasSuffix(wordlistPath, "\\") {
		// 如果wordlist路径以斜杠结尾，直接拼接
		return baseURL + "/" + wordlistPath
	}

	// 检查wordlist路径是否包含斜杠
	if strings.Contains(wordlistPath, "/") || strings.Contains(wordlistPath, "\\") {
		// 如果wordlist路径包含斜杠，直接拼接
		return baseURL + "/" + wordlistPath
	}

	// 普通情况，添加斜杠分隔
	return baseURL + "/" + wordlistPath
}

// IsDirectoryPath 判断是否为目录路径
func IsDirectoryPath(path string) bool {
	return strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\")
}

// HasSlash 检查路径是否包含斜杠
func HasSlash(path string) bool {
	return strings.Contains(path, "/") || strings.Contains(path, "\\")
}

// CleanPath 清理路径，移除多余的斜杠
func CleanPath(path string) string {
	// 移除开头的斜杠
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "\\")

	// 移除末尾的斜杠（除非是根路径）
	if path != "" {
		path = strings.TrimSuffix(path, "/")
		path = strings.TrimSuffix(path, "\\")
	}

	return path
}

// EnsureTrailingSlash 确保路径以斜杠结尾
func EnsureTrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") && !strings.HasSuffix(path, "\\") {
		return path + "/"
	}
	return path
}

// RemoveTrailingSlash 移除路径末尾的斜杠
func RemoveTrailingSlash(path string) string {
	return strings.TrimSuffix(strings.TrimSuffix(path, "/"), "\\")
}
