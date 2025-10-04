package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Map name mapping: extract map name from folder and convert to CS:GO map format
var mapNameMapping = map[string]string{
	"ancient":  "de_ancient",
	"anubis":   "de_anubis",
	"cache":    "de_cache",
	"dust2":    "de_dust2",
	"inferno":  "de_inferno",
	"mirage":   "de_mirage",
	"nuke":     "de_nuke",
	"overpass": "de_overpass",
	"train":    "de_train",
	"vertigo":  "de_vertigo",
}

type MatchFolder struct {
	Path    string
	MapName string
}

func main() {
	var outputDir string
	var targetDir string
	var dryRun bool

	flag.StringVar(&outputDir, "output", "./output", "output directory containing parsed matches")
	flag.StringVar(&targetDir, "target", "", "target CS:GO botmimic directory")
	flag.BoolVar(&dryRun, "dry", false, "dry run mode (show operations without executing)")
	flag.Parse()

	// Default target directory if not specified
	if targetDir == "" {
		targetDir = `D:\SteamLibrary\steamapps\common\Counter-Strike Global Offensive\csgo\addons\sourcemod\data\botmimic\demotest`
	}

	fmt.Println("========================================")
	fmt.Println("CS:GO Match Organizer")
	fmt.Println("========================================")
	fmt.Printf("Source directory: %s\n", outputDir)
	fmt.Printf("Target directory: %s\n", targetDir)
	if dryRun {
		fmt.Println("Mode: DRY RUN (no actual copying)")
	}
	fmt.Println("========================================\n")

	// Check if source directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		fmt.Printf("Error: Source directory does not exist: %s\n", outputDir)
		return
	}

	// Check if target directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Error: Target directory does not exist: %s\n", targetDir)
		return
	}

	// Ensure all map folders exist
	if !dryRun {
		if err := ensureMapFoldersExist(targetDir); err != nil {
			fmt.Printf("Error creating map folders: %v\n", err)
			return
		}
	}

	// Scan output directory for match folders
	matchFolders, err := scanMatchFolders(outputDir)
	if err != nil {
		fmt.Printf("Error scanning output directory: %v\n", err)
		return
	}

	if len(matchFolders) == 0 {
		fmt.Println("No match folders found in output directory")
		return
	}

	fmt.Printf("Found %d match folders\n\n", len(matchFolders))

	// Group matches by map
	matchesByMap := make(map[string][]MatchFolder)
	for _, match := range matchFolders {
		matchesByMap[match.MapName] = append(matchesByMap[match.MapName], match)
	}

	// Process each map
	totalCopied := 0
	for mapShortName, matches := range matchesByMap {
		mapFolderName := mapNameMapping[mapShortName]
		mapTargetDir := filepath.Join(targetDir, mapFolderName)

		fmt.Printf("Processing map: %s (%d matches)\n", mapFolderName, len(matches))

		// Check if map directory exists
		if _, err := os.Stat(mapTargetDir); os.IsNotExist(err) {
			fmt.Printf("  Warning: Map directory does not exist: %s\n", mapTargetDir)
			continue
		}

		// Find next available number folder
		nextNum := getNextAvailableNumber(mapTargetDir)

		// Process each match for this map
		for _, match := range matches {
			numFolderPath := filepath.Join(mapTargetDir, strconv.Itoa(nextNum))

			fmt.Printf("  [%d] %s -> %s\n", nextNum, filepath.Base(match.Path), numFolderPath)

			if !dryRun {
				// Create number folder
				if err := os.MkdirAll(numFolderPath, 0755); err != nil {
					fmt.Printf("    Error creating directory: %v\n", err)
					continue
				}

				// Copy all round folders
				roundCount, err := copyRoundFolders(match.Path, numFolderPath)
				if err != nil {
					fmt.Printf("    Error copying rounds: %v\n", err)
					continue
				}

				fmt.Printf("    Copied %d round folders\n", roundCount)
			}

			nextNum++
			totalCopied++
		}

		fmt.Println()
	}

	fmt.Println("========================================")
	if dryRun {
		fmt.Printf("Dry run complete. Would have organized %d matches\n", totalCopied)
	} else {
		fmt.Printf("Successfully organized %d matches\n", totalCopied)
	}
	fmt.Println("========================================")
}

// scanMatchFolders scans the output directory and extracts match folders with map names
func scanMatchFolders(outputDir string) ([]MatchFolder, error) {
	var matches []MatchFolder

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, err
	}

	// Regex to extract map name from folder name
	// Pattern: anything-mapname at the end
	mapPattern := regexp.MustCompile(`-([a-z0-9]+)$`)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		folderPath := filepath.Join(outputDir, folderName)

		// Extract map name from folder name
		mapName := extractMapName(folderName, mapPattern)
		if mapName == "" {
			fmt.Printf("Warning: Could not extract map name from folder: %s\n", folderName)
			continue
		}

		// Check if map is supported
		if _, exists := mapNameMapping[mapName]; !exists {
			fmt.Printf("Warning: Unsupported map '%s' in folder: %s\n", mapName, folderName)
			continue
		}

		matches = append(matches, MatchFolder{
			Path:    folderPath,
			MapName: mapName,
		})
	}

	return matches, nil
}

// extractMapName extracts map name from folder name
func extractMapName(folderName string, pattern *regexp.Regexp) string {
	// Try to match pattern like "team1-vs-team2-m1-mapname"
	matches := pattern.FindStringSubmatch(folderName)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// getNextAvailableNumber finds the next available number folder in the map directory
func getNextAvailableNumber(mapDir string) int {
	entries, err := os.ReadDir(mapDir)
	if err != nil {
		return 1
	}

	var numbers []int
	for _, entry := range entries {
		if entry.IsDir() {
			// Try to parse folder name as number
			if num, err := strconv.Atoi(entry.Name()); err == nil {
				numbers = append(numbers, num)
			}
		}
	}

	if len(numbers) == 0 {
		return 1
	}

	// Find the maximum number
	sort.Ints(numbers)
	return numbers[len(numbers)-1] + 1
}

// ensureMapFoldersExist ensures all map folders exist in target directory
func ensureMapFoldersExist(targetDir string) error {
	// List of all required map folders
	mapFolders := []string{
		"de_ancient",
		"de_anubis",
		"de_cache",
		"de_dust2",
		"de_inferno",
		"de_mirage",
		"de_nuke",
		"de_overpass",
		"de_train",
		"de_vertigo",
	}

	createdCount := 0
	for _, mapFolder := range mapFolders {
		mapPath := filepath.Join(targetDir, mapFolder)
		
		// Check if folder exists
		if _, err := os.Stat(mapPath); os.IsNotExist(err) {
			// Create the folder
			if err := os.MkdirAll(mapPath, 0755); err != nil {
				return fmt.Errorf("failed to create %s: %w", mapFolder, err)
			}
			fmt.Printf("Created map folder: %s\n", mapFolder)
			createdCount++
		}
	}

	if createdCount > 0 {
		fmt.Printf("Created %d map folders\n\n", createdCount)
	}

	return nil
}

// copyRoundFolders copies all round folders from source to destination
func copyRoundFolders(sourceDir, destDir string) (int, error) {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return 0, err
	}

	roundCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if folder starts with "round"
		if !strings.HasPrefix(entry.Name(), "round") {
			continue
		}

		sourcePath := filepath.Join(sourceDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		// Copy directory recursively
		if err := copyDir(sourcePath, destPath); err != nil {
			return roundCount, fmt.Errorf("failed to copy %s: %w", entry.Name(), err)
		}

		roundCount++
	}

	return roundCount, nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
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