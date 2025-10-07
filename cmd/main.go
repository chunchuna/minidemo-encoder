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
	var skipFreezetime bool
	
	flag.StringVar(&singleFile, "file", "", "single demo file path")
	flag.StringVar(&dirPath, "dir", "", "directory containing multiple demo files")
	flag.BoolVar(&skipFreezetime, "skipfreeze", false, "skip freezetime recording (start recording after freezetime ends)")
	flag.Parse()
	
	if dirPath != "" {
		// 批量解析模式
		parseDemoDirectory(dirPath, skipFreezetime)
	} else if singleFile != "" {
		// 单文件解析模式
		iparser.Start(singleFile, skipFreezetime)
	} else {
		fmt.Println("Usage:")
		fmt.Println("  Single file: -file=\"path/to/demo.dem\" [-skipfreeze]")
		fmt.Println("  Batch mode:  -dir=\"path/to/demo/folder\" [-skipfreeze]")
		fmt.Println("Options:")
		fmt.Println("  -skipfreeze: Skip freezetime recording, start after freezetime ends")
	}
}

func parseDemoDirectory(dirPath string, skipFreezetime bool) {
	fmt.Printf("========================================\n")
	fmt.Printf("Batch Parsing Mode\n")
	fmt.Printf("Scanning Directory: %s\n", dirPath)
	if skipFreezetime {
		fmt.Printf("Mode: Skip Freezetime\n")
	}
	fmt.Printf("========================================\n")
	
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Printf("Error: Directory does not exist: %s\n", dirPath)
		return
	}
	
	// Scan all .dem files
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
		fmt.Printf("Error: Failed to scan directory: %v\n", err)
		return
	}
	
	if len(demoFiles) == 0 {
		fmt.Printf("No .dem files found\n")
		return
	}
	
	fmt.Printf("Found %d demo files\n", len(demoFiles))
	fmt.Printf("========================================\n\n")
	
	// Parse each demo
	successCount := 0
	for i, demoPath := range demoFiles {
		// Extract demo name from filename (without extension)
		demoName := strings.TrimSuffix(filepath.Base(demoPath), filepath.Ext(demoPath))
		
		fmt.Printf("\n[%d/%d] Parsing: %s\n", i+1, len(demoFiles), demoName)
		fmt.Printf("File Path: %s\n", demoPath)
		
		// Parse demo (output subdir will be auto-set by parser to rate+demoName)
		iparser.Start(demoPath, skipFreezetime)
		
		// Reset state to avoid data confusion between multiple demos
		iencoder.ResetState()
		iparser.ResetState()
		
		successCount++
		fmt.Printf("Done Parsing: %s\n", demoName)
		fmt.Printf("----------------------------------------\n")
	}
	
	// Reset output subdirectory
	iencoder.SetOutputSubDir("")
	
	fmt.Printf("\n========================================\n")
	fmt.Printf("Batch Parsing Completed! Processed %d demo files\n", successCount)
	fmt.Printf("Output Directory: %s\n", "./output")
	fmt.Printf("========================================\n")
}
