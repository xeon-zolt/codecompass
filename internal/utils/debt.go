package utils

import (
	"bufio"
	"os"
	"regexp"

	"codecompass/internal/types"
)

// TechnicalDebtScanner provides common functionality for scanning technical debt
type TechnicalDebtScanner struct {
	todoRegex  *regexp.Regexp
	fixmeRegex *regexp.Regexp
	hackRegex  *regexp.Regexp
}

// NewTechnicalDebtScanner creates a new technical debt scanner with common patterns
func NewTechnicalDebtScanner() *TechnicalDebtScanner {
	return &TechnicalDebtScanner{
		todoRegex:  regexp.MustCompile(`(?i)//\s*todo|#\s*todo|/\*\s*todo`),
		fixmeRegex: regexp.MustCompile(`(?i)//\s*fixme|#\s*fixme|/\*\s*fixme`),
		hackRegex:  regexp.MustCompile(`(?i)//\s*hack|#\s*hack|/\*\s*hack`),
	}
}

// ScanFile scans a single file for technical debt patterns
func (tds *TechnicalDebtScanner) ScanFile(filePath string) (todoCount, fixmeCount, hackCount int, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if tds.todoRegex.MatchString(line) {
			todoCount++
		}
		if tds.fixmeRegex.MatchString(line) {
			fixmeCount++
		}
		if tds.hackRegex.MatchString(line) {
			hackCount++
		}
	}

	return todoCount, fixmeCount, hackCount, scanner.Err()
}

// ScanFiles scans multiple files and returns technical debt entries
func (tds *TechnicalDebtScanner) ScanFiles(trackedFiles map[string]bool) ([]types.TechnicalDebtEntry, error) {
	var entries []types.TechnicalDebtEntry

	for filePath := range trackedFiles {
		todoCount, fixmeCount, hackCount, err := tds.ScanFile(filePath)
		if err != nil {
			continue // Skip files that can't be read
		}

		totalDebt := todoCount + fixmeCount + hackCount
		if totalDebt > 0 {
			entries = append(entries, types.TechnicalDebtEntry{
				Path:       filePath,
				TodoCount:  todoCount,
				FixmeCount: fixmeCount,
				HackCount:  hackCount,
				TotalDebt:  totalDebt,
			})
		}
	}

	return entries, nil
}