package decipher

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dop251/goja"
)

// NSolver handles n-parameter solving using a JavaScript runtime
type NSolver struct {
	vm         *goja.Runtime
	playerCode string
	nFuncCode  string
}

// NewNSolver creates a new n-parameter solver
func NewNSolver(playerCode string) (*NSolver, error) {
	solver := &NSolver{
		vm:         goja.New(),
		playerCode: playerCode,
	}

	if err := solver.extractNFunction(); err != nil {
		return nil, fmt.Errorf("failed to extract n function: %w", err)
	}

	return solver, nil
}

// Solve solves the n-parameter challenge
func (s *NSolver) Solve(n string) (string, error) {
	if s.nFuncCode == "" {
		return n, nil
	}

	// Execute the n function in the JS runtime
	script := fmt.Sprintf(`
		%s
		nFunction("%s");
	`, s.nFuncCode, n)

	result, err := s.vm.RunString(script)
	if err != nil {
		return n, fmt.Errorf("failed to execute n function: %w", err)
	}

	if result == nil || result == goja.Undefined() || result == goja.Null() {
		return n, nil
	}

	return result.String(), nil
}

// extractNFunction extracts the n-parameter transformation function from player code
func (s *NSolver) extractNFunction() error {
	// Pattern to find the n function name
	patterns := []string{
		// Modern pattern
		`\.get\("n"\)\)&&\(b=([a-zA-Z0-9$]+)(?:\[(\d+)\])?\([a-zA-Z0-9]\)`,
		// Alternative pattern
		`\b([a-zA-Z0-9]+)\s*=\s*function\([a-zA-Z]\)\s*\{\s*var\s+[a-zA-Z]=\[[^\]]+\]`,
		// Another variant
		`(?:^|[^a-zA-Z0-9$])([a-zA-Z0-9$]+)\s*=\s*function\([a-z]\)\s*\{(?:[^}]+\}){2,}[^}]+return\s+[a-z]\.join\(""\)`,
	}

	var funcName string
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(s.playerCode)
		if len(match) >= 2 {
			funcName = match[1]
			break
		}
	}

	if funcName == "" {
		// N function not found, which is okay - some videos don't need it
		return nil
	}

	// Extract the function body
	funcBody, err := s.extractFunctionBody(funcName)
	if err != nil {
		return err
	}

	// Create wrapper for execution
	s.nFuncCode = fmt.Sprintf(`
		var nFunction = %s;
	`, funcBody)

	return nil
}

// extractFunctionBody extracts a complete function body from the player code
func (s *NSolver) extractFunctionBody(funcName string) (string, error) {
	// Escape special regex characters in function name
	escapedName := regexp.QuoteMeta(funcName)

	// Try different patterns to find the function
	patterns := []string{
		// Function expression: var name = function(...
		fmt.Sprintf(`(?:var\s+)?%s\s*=\s*(function\([^)]*\)\s*\{[^}]+(?:\{[^}]*\}[^}]*)*\})`, escapedName),
		// Function declaration: function name(...
		fmt.Sprintf(`(function\s+%s\s*\([^)]*\)\s*\{[^}]+(?:\{[^}]*\}[^}]*)*\})`, escapedName),
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(s.playerCode)
		if len(match) >= 2 {
			return match[1], nil
		}
	}

	// Try to extract using brace matching
	return s.extractFunctionWithBraceMatching(funcName)
}

// extractFunctionWithBraceMatching extracts function using brace matching
func (s *NSolver) extractFunctionWithBraceMatching(funcName string) (string, error) {
	// Find function start
	patterns := []string{
		fmt.Sprintf(`%s\s*=\s*function`, regexp.QuoteMeta(funcName)),
		fmt.Sprintf(`function\s+%s\s*\(`, regexp.QuoteMeta(funcName)),
	}

	var startIdx int = -1
	var funcStartOffset int

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		loc := re.FindStringIndex(s.playerCode)
		if loc != nil {
			startIdx = loc[0]
			// Find the actual function keyword
			funcIdx := strings.Index(s.playerCode[startIdx:], "function")
			if funcIdx >= 0 {
				funcStartOffset = funcIdx
			}
			break
		}
	}

	if startIdx < 0 {
		return "", fmt.Errorf("function %s not found", funcName)
	}

	// Find opening brace
	funcStart := startIdx + funcStartOffset
	braceStart := strings.Index(s.playerCode[funcStart:], "{")
	if braceStart < 0 {
		return "", fmt.Errorf("opening brace not found for function %s", funcName)
	}

	braceStart += funcStart + 1
	braceCount := 1
	endIdx := braceStart

	for braceCount > 0 && endIdx < len(s.playerCode) {
		switch s.playerCode[endIdx] {
		case '{':
			braceCount++
		case '}':
			braceCount--
		}
		endIdx++
	}

	if braceCount != 0 {
		return "", fmt.Errorf("unmatched braces in function %s", funcName)
	}

	return s.playerCode[funcStart:endIdx], nil
}

// BulkSolve solves multiple n-parameter challenges
func (s *NSolver) BulkSolve(challenges []string) map[string]string {
	results := make(map[string]string)

	for _, n := range challenges {
		solved, err := s.Solve(n)
		if err != nil {
			results[n] = n // Return original on error
		} else {
			results[n] = solved
		}
	}

	return results
}
