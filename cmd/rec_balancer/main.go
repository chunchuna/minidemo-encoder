package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const TARGET_REC_COUNT = 5

type TeamFolder struct {
	Path     string
	RecFiles []string
	MapName  string
	TeamType string // "t" or "ct"
}

func main() {
	var basePath string
	flag.StringVar(&basePath, "path", "", "base path to scan (e.g., D:\\SteamLibrary\\steamapps\\common\\Counter-Strike Global Offensive\\csgo\\addons\\sourcemod\\data\\botmimic\\demotest)")
	flag.Parse()

	if basePath == "" {
		fmt.Println("错误: 必须指定基础路径")
		fmt.Println("用法: rec_balancer -path=\"路径\"")
		fmt.Println("示例: rec_balancer -path=\"D:\\SteamLibrary\\steamapps\\common\\Counter-Strike Global Offensive\\csgo\\addons\\sourcemod\\data\\botmimic\\demotest\"")
		return
	}

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		fmt.Printf("错误: 路径不存在: %s\n", basePath)
		return
	}

	fmt.Println("========================================")
	fmt.Println("REC 文件平衡工具")
	fmt.Println("========================================")
	fmt.Printf("扫描路径: %s\n", basePath)
	fmt.Printf("目标数量: 每个文件夹 %d 个 REC 文件\n", TARGET_REC_COUNT)
	fmt.Println("========================================")

	// 扫描所有地图文件夹
	mapFolders, err := scanMapFolders(basePath)
	if err != nil {
		fmt.Printf("错误: 扫描失败: %v\n", err)
		return
	}

	if len(mapFolders) == 0 {
		fmt.Println("未找到任何地图文件夹")
		return
	}

	fmt.Printf("\n找到 %d 个地图文件夹\n", len(mapFolders))

	// 处理每个地图
	totalBalanced := 0
	for _, mapName := range mapFolders {
		fmt.Printf("\n----------------------------------------\n")
		fmt.Printf("处理地图: %s\n", mapName)
		fmt.Printf("----------------------------------------\n")

		mapPath := filepath.Join(basePath, mapName)
		balanced := balanceMap(mapPath, mapName)
		totalBalanced += balanced

		if balanced > 0 {
			fmt.Printf("✓ %s: 补全了 %d 个文件夹\n", mapName, balanced)
		} else {
			fmt.Printf("✓ %s: 无需补全\n", mapName)
		}
	}

	fmt.Println("\n========================================")
	fmt.Printf("平衡完成！共补全 %d 个文件夹\n", totalBalanced)
	fmt.Println("========================================")
}

// 扫描所有地图文件夹
func scanMapFolders(basePath string) ([]string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	var mapFolders []string
	for _, entry := range entries {
		if entry.IsDir() {
			// 地图文件夹通常以 de_, cs_, etc. 开头
			name := entry.Name()
			if isMapFolder(name) {
				mapFolders = append(mapFolders, name)
			}
		}
	}
	return mapFolders, nil
}

// 判断是否是地图文件夹
func isMapFolder(name string) bool {
	prefixes := []string{"de_", "cs_", "ar_", "aim_", "awp_", "fy_", "dm_", "surf_"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(name), prefix) {
			return true
		}
	}
	return false
}

// 平衡一个地图下的所有rec文件
func balanceMap(mapPath string, mapName string) int {
	// 收集所有 t 和 ct 文件夹
	tFolders := []TeamFolder{}
	ctFolders := []TeamFolder{}

	err := filepath.Walk(mapPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		dirName := strings.ToLower(filepath.Base(path))
		if dirName == "t" || dirName == "ct" {
			recFiles, _ := scanRecFiles(path)
			folder := TeamFolder{
				Path:     path,
				RecFiles: recFiles,
				MapName:  mapName,
				TeamType: dirName,
			}

			if dirName == "t" {
				tFolders = append(tFolders, folder)
			} else {
				ctFolders = append(ctFolders, folder)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("警告: 扫描地图文件夹失败: %v\n", err)
		return 0
	}

	// 平衡 t 文件夹
	tBalanced := balanceTeamFolders(tFolders, "T")
	// 平衡 ct 文件夹
	ctBalanced := balanceTeamFolders(ctFolders, "CT")

	return tBalanced + ctBalanced
}

// 扫描文件夹中的所有 .rec 文件
func scanRecFiles(folderPath string) ([]string, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var recFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".rec") {
			recFiles = append(recFiles, entry.Name())
		}
	}
	return recFiles, nil
}

// 平衡同一队伍类型的所有文件夹
func balanceTeamFolders(folders []TeamFolder, teamName string) int {
	if len(folders) == 0 {
		return 0
	}

	balancedCount := 0

	// 为每个不足的文件夹补全
	for i := range folders {
		if len(folders[i].RecFiles) >= TARGET_REC_COUNT {
			continue // 已经足够了
		}

		needed := TARGET_REC_COUNT - len(folders[i].RecFiles)
		fmt.Printf("  [%s] %s 缺少 %d 个文件\n", teamName, getRelativePath(folders[i].Path), needed)

		// 从其他文件夹收集可用的 rec 文件
		availableRecs := collectAvailableRecs(folders, i)

		if len(availableRecs) == 0 {
			fmt.Printf("    警告: 没有找到可用的 REC 文件用于补全\n")
			continue
		}

		// 复制文件
		copied := copyRecFiles(folders[i].Path, folders[i].RecFiles, availableRecs, needed)
		if copied > 0 {
			fmt.Printf("    ✓ 成功补全 %d 个文件\n", copied)
			balancedCount++
		}
	}

	return balancedCount
}

// 收集其他文件夹中可用的 rec 文件
func collectAvailableRecs(folders []TeamFolder, excludeIndex int) []string {
	existingFiles := make(map[string]bool)
	for _, filename := range folders[excludeIndex].RecFiles {
		existingFiles[strings.ToLower(filename)] = true
	}

	var available []string
	for i, folder := range folders {
		if i == excludeIndex {
			continue // 跳过当前文件夹
		}

		for _, recFile := range folder.RecFiles {
			// 检查文件是否已经存在于目标文件夹
			if !existingFiles[strings.ToLower(recFile)] {
				sourcePath := filepath.Join(folder.Path, recFile)
				available = append(available, sourcePath)
				existingFiles[strings.ToLower(recFile)] = true
			}
		}
	}

	return available
}

// 复制 rec 文件到目标文件夹
func copyRecFiles(targetFolder string, existingFiles []string, availableRecs []string, needed int) int {
	copied := 0

	for i := 0; i < len(availableRecs) && copied < needed; i++ {
		sourcePath := availableRecs[i]
		fileName := filepath.Base(sourcePath)

		// 检查文件名是否已存在（不区分大小写）
		exists := false
		for _, existing := range existingFiles {
			if strings.EqualFold(existing, fileName) {
				exists = true
				break
			}
		}

		if exists {
			continue
		}

		targetPath := filepath.Join(targetFolder, fileName)

		// 如果目标文件已存在，生成新的文件名
		if _, err := os.Stat(targetPath); err == nil {
			// 文件已存在，添加后缀
			ext := filepath.Ext(fileName)
			nameWithoutExt := strings.TrimSuffix(fileName, ext)
			for j := 1; ; j++ {
				newFileName := fmt.Sprintf("%s_%d%s", nameWithoutExt, j, ext)
				targetPath = filepath.Join(targetFolder, newFileName)
				if _, err := os.Stat(targetPath); os.IsNotExist(err) {
					break
				}
			}
		}

		// 复制文件
		if err := copyFile(sourcePath, targetPath); err != nil {
			fmt.Printf("    警告: 复制失败 %s: %v\n", fileName, err)
			continue
		}

		fmt.Printf("    → 复制: %s\n", fileName)
		copied++
	}

	return copied
}

// 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// 获取相对路径用于显示
func getRelativePath(fullPath string) string {
	parts := strings.Split(filepath.ToSlash(fullPath), "/")
	if len(parts) >= 3 {
		return filepath.Join(parts[len(parts)-3:]...)
	}
	return fullPath
} 