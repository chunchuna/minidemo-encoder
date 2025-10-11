package main

import (
	"flag"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func main() {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 解析命令行参数
	startDir := flag.String("dir", ".", "起始目录路径")
	flag.Parse()

	fmt.Printf("开始在 %s 目录及其子目录中查找 .rec 文件...\n\n", *startDir)

	var (
		filesFound    = 0
		filesRenamed  = 0
		errorCount    = 0
		wg            sync.WaitGroup
		mutex         sync.Mutex
		throttle      = make(chan struct{}, 10) // 限制最多10个并发操作
	)

	// 遍历目录
	err := filepath.WalkDir(*startDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("访问路径出错 %s: %v\n", path, err)
			return filepath.SkipDir
		}

		// 检查是否为.rec文件
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".rec") {
			filesFound++
			
			// 并发处理文件重命名
			wg.Add(1)
			throttle <- struct{}{} // 获取信号量，限制并发
			
			go func(filePath string) {
				defer wg.Done()
				defer func() { <-throttle }() // 释放信号量
				
				// 生成8位随机数字
				randomName := generateRandomName(8)
				
				// 构造新文件名
				dir := filepath.Dir(filePath)
				oldName := filepath.Base(filePath)
				newName := randomName + "player.rec"
				newPath := filepath.Join(dir, newName)
				
				// 重命名文件
				if err := os.Rename(filePath, newPath); err != nil {
					mutex.Lock()
					errorCount++
					fmt.Printf("重命名失败: %s -> %s: %v\n", filePath, newName, err)
					mutex.Unlock()
				} else {
					mutex.Lock()
					filesRenamed++
					fmt.Printf("成功: %s -> %s\n", oldName, newName)
					mutex.Unlock()
				}
			}(path)
		}
		return nil
	})

	// 等待所有重命名操作完成
	wg.Wait()
	close(throttle)

	if err != nil {
		fmt.Printf("\n遍历目录时出错: %v\n", err)
	}

	// 打印统计信息
	fmt.Printf("\n完成!\n")
	fmt.Printf("找到的 .rec 文件总数: %d\n", filesFound)
	fmt.Printf("成功重命名的文件: %d\n", filesRenamed)
	if errorCount > 0 {
		fmt.Printf("重命名失败的文件: %d\n", errorCount)
	}
}

// 生成指定长度的随机数字字符串
func generateRandomName(length int) string {
	digits := make([]byte, length)
	for i := 0; i < length; i++ {
		digits[i] = byte('0' + rand.Intn(10))
	}
	return string(digits)
} 