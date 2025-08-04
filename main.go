package main

import (
	"fmt"
	"os"

	"dirsearch-go/internal/cmd"
	"dirsearch-go/internal/config"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	// 运行命令行程序
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
