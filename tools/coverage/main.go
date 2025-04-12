package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type coverage struct {
	file      string
	lineInfo  map[int]bool // line number -> covered (true) or not (false)
	blockInfo map[int]block // start line -> block info
}

type block struct {
	startLine int
	endLine   int
	isIf      bool
	covered   int // 0 = not covered, 1 = partially covered, 2 = fully covered
}

func main() {
	// Parse command line arguments
	var coverageFile string
	var targetFile string
	var showFullFile bool
	var highlightConditionals bool
	var coveredSymbol string
	var uncoveredSymbol string
	var untrackedSymbol string
	var noEmoji bool

	flag.StringVar(&coverageFile, "cov", "coverage.out", "Path to coverage.out file")
	flag.StringVar(&targetFile, "file", "", "Specific file to show coverage for (optional)")
	flag.BoolVar(&showFullFile, "full", true, "Show all lines (true) or only lines with coverage data (false)")
	flag.BoolVar(&highlightConditionals, "highlight-if", true, "Highlight if statements and their coverage status")
	flag.StringVar(&coveredSymbol, "covered", "✅", "Symbol for covered lines")
	flag.StringVar(&uncoveredSymbol, "uncovered", "❌", "Symbol for uncovered lines")
	flag.StringVar(&untrackedSymbol, "untracked", "⚪", "Symbol for lines not tracked in coverage")
	flag.BoolVar(&noEmoji, "no-emoji", false, "Use text instead of emoji (useful for terminals without emoji support)")
	flag.Parse()

	// If noEmoji flag is set, use text symbols instead
	if noEmoji {
		coveredSymbol = "COVERED"
		uncoveredSymbol = "NOT-COV"
		untrackedSymbol = "IGNORED"
	}
	
	// Read and parse coverage data
	coverageData := make(map[string]*coverage)
	if err := parseCoverageFile(coverageFile, coverageData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing coverage file: %v\n", err)
		os.Exit(1)
	}

	// If a specific file is requested, only show that file
	if targetFile != "" {
		// Find the exact match or matches that contain this path
		var matches []string
		for covFile := range coverageData {
			if strings.Contains(covFile, targetFile) {
				matches = append(matches, covFile)
			}
		}

		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "No coverage data found for file: %s\n", targetFile)
			os.Exit(1)
		} else if len(matches) > 1 {
			fmt.Println("Multiple matching files found:")
			for _, match := range matches {
				fmt.Println("  ", match)
			}
			fmt.Println("Please specify a more precise file path.")
			os.Exit(1)
		}

		// Process conditional blocks if enabled
		if highlightConditionals {
			analyzeConditionals(matches[0], coverageData[matches[0]])
		}

		// Show coverage for the matched file
		if err := displayFileCoverage(matches[0], coverageData[matches[0]], showFullFile, highlightConditionals, 
			coveredSymbol, uncoveredSymbol, untrackedSymbol, noEmoji); err != nil {
			fmt.Fprintf(os.Stderr, "Error displaying coverage: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Show all files with coverage data
		fmt.Println("Available files with coverage data:")
		for file := range coverageData {
			fmt.Println("  ", file)
		}
		fmt.Println("\nRun with -file flag to show coverage for a specific file.")
	}
}

func parseCoverageFile(filePath string, result map[string]*coverage) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Skip the first line which is usually "mode: set"
	if scanner.Scan() && !strings.HasPrefix(scanner.Text(), "github.com") {
		// First line was the mode, continue
	}

	// Regular expression to parse coverage data lines
	// Format: github.com/user-story-matrix/usm/internal/workflow/executor.go:23.69,28.2 1 1
	lineRegex := regexp.MustCompile(`^(.+):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)$`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := lineRegex.FindStringSubmatch(line)
		if len(matches) < 8 {
			continue // Skip invalid lines
		}

		filePath := matches[1]
		startLine, _ := strconv.Atoi(matches[2])
		endLine, _ := strconv.Atoi(matches[4])
		numStmt, _ := strconv.Atoi(matches[6])
		count, _ := strconv.Atoi(matches[7])

		covered := count > 0

		// Initialize coverage data for this file if it doesn't exist
		if _, exists := result[filePath]; !exists {
			result[filePath] = &coverage{
				file:      filePath,
				lineInfo:  make(map[int]bool),
				blockInfo: make(map[int]block),
			}
		}

		// Add line coverage information
		for line := startLine; line <= endLine; line++ {
			// If multiple statements share coverage information, numStmt > 1
			// For simplicity, we'll mark all lines in the range with the same coverage status
			if numStmt > 0 {
				result[filePath].lineInfo[line] = covered
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func analyzeConditionals(filePath string, cov *coverage) error {
	// Find the local path to the file
	localPath, err := findLocalPath(filePath)
	if err != nil {
		return err
	}

	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Regular expression to find if statements
	ifRegex := regexp.MustCompile(`^\s*if\s+.*{.*$`)
	
	// Scan through the file
	scanner := bufio.NewScanner(file)
	lineNum := 1
	var ifBlocks []block
	var currentBlock *block
	var openBraces int

	for scanner.Scan() {
		line := scanner.Text()
		
		// Check if this is an if statement
		if ifRegex.MatchString(line) {
			// Create a new if block
			currentBlock = &block{
				startLine: lineNum,
				endLine:   -1, // Will be filled later
				isIf:      true,
				covered:   0,  // Will be calculated later
			}
			
			// Count opening braces
			openBraces = strings.Count(line, "{") - strings.Count(line, "}")
			
			// If openBraces is 0, this is a single-line if, so just set the end line to the same line
			if openBraces == 0 {
				currentBlock.endLine = lineNum
				ifBlocks = append(ifBlocks, *currentBlock)
				currentBlock = nil
			}
		} else if currentBlock != nil {
			// Update brace count
			openBraces += strings.Count(line, "{") - strings.Count(line, "}")
			
			// If braces are balanced, we've reached the end of the block
			if openBraces <= 0 {
				currentBlock.endLine = lineNum
				ifBlocks = append(ifBlocks, *currentBlock)
				currentBlock = nil
			}
		}
		
		lineNum++
	}

	// If there are any unfinished blocks, finalize them
	if currentBlock != nil {
		currentBlock.endLine = lineNum - 1
		ifBlocks = append(ifBlocks, *currentBlock)
	}

	// Calculate coverage for each if block
	for i, block := range ifBlocks {
		if block.endLine == -1 {
			// If we couldn't find the end, just set it to the start + 1 as a safe estimate
			ifBlocks[i].endLine = block.startLine + 1
		}
		
		// First, check if the if statement line itself is in the coverage data
		conditionCovered, conditionExists := cov.lineInfo[block.startLine]
		
		// Check if all lines in the if block are covered
		allCovered := true
		hasData := false
		
		// Then check the block body (only if different from the if line)
		for line := block.startLine; line <= block.endLine; line++ {
			covered, exists := cov.lineInfo[line]
			if exists {
				hasData = true
				if !covered {
					allCovered = false
				}
			}
		}
		
		// Set coverage status:
		// - Fully covered (2): Condition and all statements are covered
		// - Partially covered (1): Condition exists in coverage data but not all paths taken
		// - Not covered (0): No coverage data for this block
		
		if conditionExists && conditionCovered && allCovered && hasData {
			// The condition and all statements are covered - fully covered
			ifBlocks[i].covered = 2
		} else if conditionExists {
			// The condition exists in coverage data - at least partially covered
			// Even if marked as not covered, it was evaluated but never true
			ifBlocks[i].covered = 1
		} else if hasData {
			// Some lines in the block have coverage data but condition doesn't - odd case
			ifBlocks[i].covered = 1
		} else {
			// No coverage data at all
			ifBlocks[i].covered = 0
		}
		
		// Store the if block in the coverage data
		cov.blockInfo[block.startLine] = ifBlocks[i]
	}

	return nil
}

func displayFileCoverage(filePath string, cov *coverage, showFullFile bool, highlightConditionals bool,
	coveredSymbol, uncoveredSymbol, untrackedSymbol string, noEmoji bool) error {
	// For Go modules, we need to convert the module path to a file path
	localPath, err := findLocalPath(filePath)
	if err != nil {
		return err
	}

	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1

	fmt.Printf("\nCoverage for %s:\n\n", filePath)

	// Find the maximum line number to determine padding
	maxLine := 0
	// Count total lines in file
	totalLines := 0
	fileLines := []string{}
	for scanner.Scan() {
		totalLines++
		fileLines = append(fileLines, scanner.Text())
	}
	
	// Use total lines for padding or the highest line in the coverage data
	maxLine = totalLines
	for line := range cov.lineInfo {
		if line > maxLine {
			maxLine = line
		}
	}
	padding := len(strconv.Itoa(maxLine))

	// Reset the file scanner
	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)
	lineNum = 1

	// Display all lines with their coverage status
	for scanner.Scan() {
		line := scanner.Text()
		var symbol string
		var prefix string
		
		if covered, exists := cov.lineInfo[lineNum]; exists {
			if covered {
				symbol = coveredSymbol
			} else {
				symbol = uncoveredSymbol
			}
			
			// Check if this is part of an if block and highlight it
			if highlightConditionals {
				for _, block := range cov.blockInfo {
					if block.isIf && lineNum >= block.startLine && lineNum <= block.endLine {
						if lineNum == block.startLine {
							switch block.covered {
							case 2:
								// Fully covered
								if noEmoji {
									prefix = "[FULL] "
								} else {
									prefix = "🟢 "
								}
							case 1:
								// Partially covered
								if noEmoji {
									prefix = "[PART] "
								} else {
									prefix = "🟡 "
								}
							case 0:
								// Not covered
								if noEmoji {
									prefix = "[NONE] "
								} else {
									prefix = "🔴 "
								}
							}
						}
						break
					}
				}
			}
			
			fmt.Printf("%*d %s | %s%s\n", padding, lineNum, symbol, prefix, line)
		} else if showFullFile {
			// Use a different symbol for lines not tracked in coverage
			symbol = untrackedSymbol

			// Still highlight if blocks even for untracked lines
			if highlightConditionals {
				block, exists := cov.blockInfo[lineNum]
				if exists && block.isIf {
					switch block.covered {
					case 2:
						// Fully covered
						if noEmoji {
							prefix = "[FULL] "
						} else {
							prefix = "🟢 "
						}
					case 1:
						// Partially covered
						if noEmoji {
							prefix = "[PART] "
						} else {
							prefix = "🟡 "
						}
					case 0:
						// Not covered
						if noEmoji {
							prefix = "[NONE] "
						} else {
							prefix = "🔴 "
						}
					}
				}
			}
			
			fmt.Printf("%*d %s | %s%s\n", padding, lineNum, symbol, prefix, line)
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	
	if highlightConditionals {
		fmt.Println("\nConditional Coverage Legend:")
		if noEmoji {
			fmt.Println("[FULL] Fully covered if statement")
			fmt.Println("[PART] Partially covered if statement (some branches may not be covered)")
			fmt.Println("[NONE] Uncovered if statement")
		} else {
			fmt.Println("🟢 Fully covered if statement")
			fmt.Println("🟡 Partially covered if statement (some branches may not be covered)")
			fmt.Println("🔴 Uncovered if statement")
		}
	}

	return nil
}

func findLocalPath(modulePath string) (string, error) {
	// Try to find the file directly
	if _, err := os.Stat(modulePath); err == nil {
		return modulePath, nil
	}

	// Try to find the file by stripping the module prefix
	parts := strings.Split(modulePath, "/")
	for i := 0; i < len(parts); i++ {
		testPath := filepath.Join(parts[i:]...)
		if _, err := os.Stat(testPath); err == nil {
			return testPath, nil
		}
	}

	// If we have the GOPATH, we could look there too
	gopath := os.Getenv("GOPATH")
	if gopath != "" {
		srcPath := filepath.Join(gopath, "src", modulePath)
		if _, err := os.Stat(srcPath); err == nil {
			return srcPath, nil
		}
	}

	// As a last resort, use find to locate the file in the current directory
	cmd := fmt.Sprintf("find . -path '*%s' 2>/dev/null", filepath.Base(modulePath))
	findOutput, err := runCommand(cmd)
	if err == nil && findOutput != "" {
		lines := strings.Split(findOutput, "\n")
		for _, line := range lines {
			if line != "" && strings.HasSuffix(line, filepath.Base(modulePath)) {
				return line, nil
			}
		}
	}

	return "", fmt.Errorf("could not locate source file for %s", modulePath)
}

func runCommand(cmd string) (string, error) {
	command := exec.Command("sh", "-c", cmd)
	output, err := command.Output()
	return strings.TrimSpace(string(output)), err
} 