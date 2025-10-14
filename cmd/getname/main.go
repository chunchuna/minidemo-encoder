package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("获取当前目录失败:", err)
		return
	}

	// 存储找到的rec文件名（不带后缀）
	recFiles := make(map[string]bool)
	var fileNames []string

	// 定义一个错误来停止遍历
	var stopWalk = errors.New("找到足够的文件")

	// 遍历目录
	err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 如果已经找到10个，停止遍历
		if len(fileNames) >= 10 {
			return stopWalk
		}

		// 检查是否是rec文件
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".rec") {
			// 获取不带后缀的文件名
			nameWithoutExt := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			
			// 检查是否重复
			if !recFiles[nameWithoutExt] {
				recFiles[nameWithoutExt] = true
				fileNames = append(fileNames, nameWithoutExt)
			}
		}

		return nil
	})

	if err != nil && err != stopWalk {
		fmt.Println("遍历目录失败:", err)
		return
	}

	if len(fileNames) == 0 {
		fmt.Println("未找到任何rec文件")
		return
	}

	// 创建name.txt文件
	outputFile := filepath.Join(currentDir, "name.txt")
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("创建name.txt失败:", err)
		return
	}
	defer file.Close()

	// 写入文件名
	for _, name := range fileNames {
		_, err := file.WriteString(name + "\n")
		if err != nil {
			fmt.Println("写入文件失败:", err)
			return
		}
	}

	fmt.Printf("成功找到 %d 个rec文件，已写入name.txt\n", len(fileNames))
}
