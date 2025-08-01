package coverage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"codecompass/internal/types"
)

// DetectCoverageFile attempts to find coverage files in common locations
func DetectCoverageFile() (string, error) {
	commonPaths := []string{
		"coverage/lcov.info",
		"coverage/coverage.info",
		"lcov.info",
		"coverage.info",
		"coverage/coverage-final.json",
		"coverage-final.json",
		"nyc_output/coverage-final.json",
		".nyc_output/coverage-final.json",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath, nil
		}
	}

	return "", fmt.Errorf("no coverage file found in common locations")
}

// ParseCoverageFile parses different types of coverage files
func ParseCoverageFile(filePath string) (*types.CoverageData, error) {
	if filePath == "" {
		detectedPath, err := DetectCoverageFile()
		if err != nil {
			return nil, err
		}
		filePath = detectedPath
	}

	// Determine file type by extension
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".info":
		return parseLcovFile(filePath)
	case ".json":
		return parseJsonCoverageFile(filePath)
	default:
		// Try to auto-detect by content
		return parseAutoDetect(filePath)
	}
}

// parseLcovFile parses LCOV info files
func parseLcovFile(filePath string) (*types.CoverageData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	coverage := &types.CoverageData{
		Files: make(map[string]types.FileCoverage),
	}

	scanner := bufio.NewScanner(file)
	var currentFile types.FileCoverage
	var currentPath string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "SF:") {
			// Source file
			currentPath = strings.TrimPrefix(line, "SF:")
			currentFile = types.FileCoverage{Path: currentPath}
		} else if strings.HasPrefix(line, "LH:") {
			// Lines hit
			if count, err := strconv.Atoi(strings.TrimPrefix(line, "LH:")); err == nil {
				currentFile.LinesCovered = count
			}
		} else if strings.HasPrefix(line, "LF:") {
			// Lines found
			if count, err := strconv.Atoi(strings.TrimPrefix(line, "LF:")); err == nil {
				currentFile.LinesTotal = count
			}
		} else if strings.HasPrefix(line, "FNH:") {
			// Functions hit
			if count, err := strconv.Atoi(strings.TrimPrefix(line, "FNH:")); err == nil {
				currentFile.FunctionsCovered = count
			}
		} else if strings.HasPrefix(line, "FNF:") {
			// Functions found
			if count, err := strconv.Atoi(strings.TrimPrefix(line, "FNF:")); err == nil {
				currentFile.FunctionsTotal = count
			}
		} else if strings.HasPrefix(line, "BRH:") {
			// Branches hit
			if count, err := strconv.Atoi(strings.TrimPrefix(line, "BRH:")); err == nil {
				currentFile.BranchesCovered = count
			}
		} else if strings.HasPrefix(line, "BRF:") {
			// Branches found
			if count, err := strconv.Atoi(strings.TrimPrefix(line, "BRF:")); err == nil {
				currentFile.BranchesTotal = count
			}
		} else if line == "end_of_record" {
			// End of current file record
			if currentPath != "" {
				coverage.Files[currentPath] = currentFile
			}
		}
	}

	return coverage, scanner.Err()
}

// parseJsonCoverageFile parses JSON coverage files (NYC/Istanbul format)
func parseJsonCoverageFile(filePath string) (*types.CoverageData, error) {
	// This would implement JSON parsing for NYC/Istanbul coverage
	// For now, return an error to indicate it's not implemented
	return nil, fmt.Errorf("JSON coverage file parsing not yet implemented")
}

// parseAutoDetect attempts to auto-detect file format
func parseAutoDetect(filePath string) (*types.CoverageData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		firstLine := scanner.Text()
		if strings.HasPrefix(firstLine, "TN:") || strings.HasPrefix(firstLine, "SF:") {
			// Looks like LCOV format
			file.Seek(0, 0) // Reset file pointer
			file.Close()
			return parseLcovFile(filePath)
		} else if strings.HasPrefix(firstLine, "{") {
			// Looks like JSON
			file.Close()
			return parseJsonCoverageFile(filePath)
		}
	}

	return nil, fmt.Errorf("unable to detect coverage file format")
}

// GetCoverageStats calculates coverage statistics
func GetCoverageStats(coverage *types.CoverageData, trackedFiles map[string]bool) []types.CoverageEntry {
	var entries []types.CoverageEntry

	for filePath, fileCoverage := range coverage.Files {
		// Convert absolute path to relative if needed
		relPath := filePath
		if filepath.IsAbs(filePath) {
			if rel, err := filepath.Rel(".", filePath); err == nil {
				relPath = rel
			}
		}

		// Only include tracked files
		if !trackedFiles[relPath] && !trackedFiles[filePath] {
			continue
		}

		var coveragePercent float64
		if fileCoverage.LinesTotal > 0 {
			coveragePercent = float64(fileCoverage.LinesCovered) / float64(fileCoverage.LinesTotal) * 100
		}

		entries = append(entries, types.CoverageEntry{
			Path:             relPath,
			LinesCovered:     fileCoverage.LinesCovered,
			LinesTotal:       fileCoverage.LinesTotal,
			CoveragePercent:  coveragePercent,
			FunctionsCovered: fileCoverage.FunctionsCovered,
			FunctionsTotal:   fileCoverage.FunctionsTotal,
			BranchesCovered:  fileCoverage.BranchesCovered,
			BranchesTotal:    fileCoverage.BranchesTotal,
		})
	}

	return entries
}
