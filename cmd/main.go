package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	iparser "github.com/hx-w/minidemo-encoder/internal/parser"
	iencoder "github.com/hx-w/minidemo-encoder/internal/encoder"
)

func main() {
	var singleFile string
	var dirPath string
	
	flag.StringVar(&singleFile, "file", "", "single demo file path")
	flag.StringVar(&dirPath, "dir", "", "directory containing multiple demo files")
	flag.Parse()
	
	if dirPath != "" {
		// 批量解析模式
		parseDemoDirectory(dirPath)
	} else if singleFile != "" {
		// 单文件解析模式
		iparser.Start(singleFile)
	} else {
		fmt.Println("Usage:")
		fmt.Println("  Single file: -file=\"path/to/demo.dem\"")
		fmt.Println("  Batch mode:  -dir=\"path/to/demo/folder\"")
	}
}

func parseDemoDirectory(dirPath string) {
	fmt.Printf("========================================\n")
	fmt.Printf("批量解析模式\n")
	fmt.Printf("扫描目录: %s\n", dirPath)
	fmt.Printf("========================================\n")
	
	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Printf("错误: 目录不存在: %s\n", dirPath)
		return
	}
	
	// 扫描所有 .dem 文件
	demoFiles := []string{}
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".dem") {
			demoFiles = append(demoFiles, path)
		}
		return nil
	})
	
	if err != nil {
		fmt.Printf("错误: 扫描目录失败: %v\n", err)
		return
	}
	
	if len(demoFiles) == 0 {
		fmt.Printf("未找到任何 .dem 文件\n")
		return
	}
	
	fmt.Printf("找到 %d 个 demo 文件\n", len(demoFiles))
	fmt.Printf("========================================\n\n")
	
	// 解析每个 demo
	successCount := 0
	for i, demoPath := range demoFiles {
		// 从文件名提取 demo 名称（不含扩展名）
		demoName := strings.TrimSuffix(filepath.Base(demoPath), filepath.Ext(demoPath))
		
		fmt.Printf("\n[%d/%d] 正在解析: %s\n", i+1, len(demoFiles), demoName)
		fmt.Printf("文件路径: %s\n", demoPath)
		
		// 设置输出子目录为 demo 名称
		iencoder.SetOutputSubDir(demoName)
		
		// 解析 demo
		iparser.Start(demoPath)
		
		// 重置状态，避免多个demo之间数据混乱
		iencoder.ResetState()
		
		successCount++
		fmt.Printf("✓ 完成解析: %s\n", demoName)
		fmt.Printf("----------------------------------------\n")
	}
	
	// 重置输出子目录
	iencoder.SetOutputSubDir("")
	
	fmt.Printf("\n========================================\n")
	fmt.Printf("批量解析完成！共处理 %d 个 demo 文件\n", successCount)
	fmt.Printf("输出目录: %s\n", "./output")
	fmt.Printf("========================================\n")
}
