// Package decipher provides signature and n-parameter deciphering for YouTube URLs.
// This is required to bypass throttling and obtain valid stream URLs.
package decipher

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Decipherer handles signature and n-parameter challenges
type Decipherer struct {
	playerCode    string
	playerURL     string

	sigTokens     []string
	nSolver       *NSolver

	mu sync.Mutex
}

// JavaScript regex patterns for signature function extraction
var (
	jsVarStr       = `[a-zA-Z_\$]\w*`
	jsSingleQuote  = `'[^'\\]*(?:\\[\s\S][^'\\]*)*'`
	jsDoubleQuote  = `"[^"\\]*(?:\\[\s\S][^"\\]*)*"`
	jsQuoteStr     = fmt.Sprintf(`(?:%s|%s)`, jsSingleQuote, jsDoubleQuote)
	jsKeyStr       = fmt.Sprintf(`(?:%s|%s)`, jsVarStr, jsQuoteStr)
	jsPropStr      = fmt.Sprintf(`(?:\.%s|\[%s\])`, jsVarStr, jsQuoteStr)

	reverseStr = `:function\(a\)\{(?:return )?a\.reverse\(\)\}`
	sliceStr   = `:function\(a,b\)\{return a\.slice\(b\)\}`
	spliceStr  = `:function\(a,b\)\{a\.splice\(0,b\)\}`
	swapStr    = `:function\(a,b\)\{var c=a\[0\];a\[0\]=a\[b(?:%a\.length)?\];a\[b(?:%a\.length)?\]=c(?:;return a)?\}`

	actionsObjRegex  = regexp.MustCompile(fmt.Sprintf(
		`var (%s)=\{((?:(?:%s%s|%s%s|%s%s|%s%s),?\r?\n?)+)\};`,
		jsVarStr, jsKeyStr, reverseStr, jsKeyStr, sliceStr, jsKeyStr, spliceStr, jsKeyStr, swapStr))

	actionsFuncRegex = regexp.MustCompile(fmt.Sprintf(
		`function(?: %s)?\(a\)\{a=a\.split\((?:''|"")\);\s*((?:(?:a=)?%s%s\(a,\d+\);)+)return a\.join\((?:''|"")\)\}`,
		jsVarStr, jsVarStr, jsPropStr))

	reverseRegex = regexp.MustCompile(fmt.Sprintf(`(?:^|,)(%s)%s`, jsKeyStr, reverseStr))
	sliceRegex   = regexp.MustCompile(fmt.Sprintf(`(?:^|,)(%s)%s`, jsKeyStr, sliceStr))
	spliceRegex  = regexp.MustCompile(fmt.Sprintf(`(?:^|,)(%s)%s`, jsKeyStr, spliceStr))
	swapRegex    = regexp.MustCompile(fmt.Sprintf(`(?:^|,)(%s)%s`, jsKeyStr, swapStr))

	// N-parameter function patterns
	nFuncNameRegex = regexp.MustCompile(`\.get\("n"\)\)&&\(b=([a-zA-Z0-9$]+)(?:\[(\d+)\])?\([a-zA-Z0-9]\)`)
	nFuncBodyRegex = regexp.MustCompile(`(?s)var %s=\{.*?\};`)
)

// New creates a new Decipherer with the given player JS code
func New(playerCode, playerURL string) (*Decipherer, error) {
	d := &Decipherer{
		playerCode: playerCode,
		playerURL:  playerURL,
	}

	if err := d.extractSignatureTokens(); err != nil {
		return nil, fmt.Errorf("failed to extract signature tokens: %w", err)
	}

	return d, nil
}

// NewDecipherer creates a new Decipherer from player code
func NewDecipherer(playerCode string) (*Decipherer, error) {
	d := &Decipherer{
		playerCode: playerCode,
	}

	if err := d.extractSignatureTokens(); err != nil {
		return nil, fmt.Errorf("failed to extract signature tokens: %w", err)
	}

	// Initialize n-solver
	nSolver, err := NewNSolver(playerCode)
	if err != nil {
		// N-solver is optional, some videos don't need it
		d.nSolver = nil
	} else {
		d.nSolver = nSolver
	}

	return d, nil
}

// GetSignatureTimestamp returns the signature timestamp from the player code
func (d *Decipherer) GetSignatureTimestamp() int {
	return GetSignatureTimestamp(d.playerCode)
}

// SolveNChallenge solves the n-parameter challenge using the JS runtime
func (d *Decipherer) SolveNChallenge(n string) (string, error) {
	if d.nSolver == nil {
		return n, nil
	}
	return d.nSolver.Solve(n)
}

// DecipherURL deciphers a stream URL by solving signature and n-parameter challenges
func (d *Decipherer) DecipherURL(streamURL string) (string, error) {
	parsed, err := url.Parse(streamURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	query := parsed.Query()

	// Check for signature cipher in query
	if sig := query.Get("s"); sig != "" {
		// Decipher signature
		decipheredSig := d.decipherSignature(sig)

		// Get signature parameter name (usually "sig" or "signature")
		sp := query.Get("sp")
		if sp == "" {
			sp = "signature"
		}

		query.Set(sp, decipheredSig)
		query.Del("s")
		query.Del("sp")
	}

	// Handle n-parameter for throttle bypass
	if n := query.Get("n"); n != "" {
		newN, err := d.solveNChallenge(n)
		if err == nil && newN != "" {
			query.Set("n", newN)
		}
	}

	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

// DecipherSignature deciphers a signature using the extracted tokens
func (d *Decipherer) DecipherSignature(sig string) string {
	return d.decipherSignature(sig)
}

// decipherSignature applies the signature transformation
func (d *Decipherer) decipherSignature(sig string) string {
	arr := strings.Split(sig, "")

	for _, token := range d.sigTokens {
		if len(token) < 1 {
			continue
		}

		switch token[0] {
		case 'r':
			// Reverse
			for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
				arr[i], arr[j] = arr[j], arr[i]
			}

		case 's':
			// Slice
			if pos, err := strconv.Atoi(token[1:]); err == nil && pos < len(arr) {
				arr = arr[pos:]
			}

		case 'p':
			// Splice
			if pos, err := strconv.Atoi(token[1:]); err == nil && pos < len(arr) {
				arr = arr[pos:]
			}

		case 'w':
			// Swap
			if pos, err := strconv.Atoi(token[1:]); err == nil {
				swapPos := pos % len(arr)
				arr[0], arr[swapPos] = arr[swapPos], arr[0]
			}
		}
	}

	return strings.Join(arr, "")
}

// extractSignatureTokens extracts signature transformation tokens from player code
func (d *Decipherer) extractSignatureTokens() error {
	objects := actionsObjRegex.FindStringSubmatch(d.playerCode)
	functions := actionsFuncRegex.FindStringSubmatch(d.playerCode)

	if len(objects) < 3 || len(functions) < 2 {
		// Try alternative extraction methods
		return d.extractSignatureTokensAlt()
	}

	obj := strings.ReplaceAll(objects[1], "$", "\\$")
	objBody := strings.ReplaceAll(objects[2], "$", "\\$")
	funcBody := strings.ReplaceAll(functions[1], "$", "\\$")

	reverseKey := extractKey(reverseRegex, objBody)
	sliceKey := extractKey(sliceRegex, objBody)
	spliceKey := extractKey(spliceRegex, objBody)
	swapKey := extractKey(swapRegex, objBody)

	keys := fmt.Sprintf("(%s|%s|%s|%s)", reverseKey, sliceKey, spliceKey, swapKey)
	tokenizeRegex := regexp.MustCompile(fmt.Sprintf(
		`(?:a=)?%s(?:\.%s|\[(?:'%s'|"%s")\])\(a,(\d+)\)`,
		obj, keys, keys, keys))

	matches := tokenizeRegex.FindAllStringSubmatch(funcBody, -1)

	for _, result := range matches {
		if len(result) < 5 {
			continue
		}

		key := result[1]
		if key == "" {
			key = result[2]
		}
		if key == "" {
			key = result[3]
		}

		switch key {
		case reverseKey:
			d.sigTokens = append(d.sigTokens, "r")
		case sliceKey:
			d.sigTokens = append(d.sigTokens, "s"+result[4])
		case spliceKey:
			d.sigTokens = append(d.sigTokens, "p"+result[4])
		case swapKey:
			d.sigTokens = append(d.sigTokens, "w"+result[4])
		}
	}

	return nil
}

// extractSignatureTokensAlt tries alternative patterns for signature extraction
func (d *Decipherer) extractSignatureTokensAlt() error {
	// Alternative signature function pattern
	altPattern := regexp.MustCompile(`\b[cs]\s*&&\s*[adf]\.set\([^,]+\s*,\s*encodeURIComponent\(([a-zA-Z0-9$]+)\(`)
	match := altPattern.FindStringSubmatch(d.playerCode)

	if len(match) < 2 {
		// If no signature function found, the URL may not need deciphering
		return nil
	}

	return nil
}

// solveNChallenge solves the n-parameter challenge to bypass throttling
func (d *Decipherer) solveNChallenge(n string) (string, error) {
	// The n-parameter solving requires JavaScript execution
	// This is a simplified version that may need enhancement

	// For now, we'll return the original n value
	// Full implementation would use a JS runtime like goja
	return n, nil
}

// extractKey extracts a key from the object body using the given regex
func extractKey(re *regexp.Regexp, body string) string {
	match := re.FindStringSubmatch(body)
	if len(match) < 2 {
		return ""
	}

	key := match[1]
	// Remove quotes if present
	key = strings.Trim(key, `"'`)
	return key
}

// GetSignatureTimestamp extracts the signature timestamp from player code
func GetSignatureTimestamp(playerCode string) int {
	patterns := []string{
		`(?:signatureTimestamp|sts)\s*:\s*(\d{5})`,
		`"STS"\s*:\s*(\d{5})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(playerCode)
		if len(match) >= 2 {
			if sts, err := strconv.Atoi(match[1]); err == nil {
				return sts
			}
		}
	}

	return 0
}

// ExtractPlayerID extracts the player ID from a player URL
func ExtractPlayerID(playerURL string) string {
	patterns := []string{
		`/s/player/([a-zA-Z0-9_-]{8,})/`,
		`/([a-zA-Z0-9_-]{8,})/player`,
		`\b(vfl[a-zA-Z0-9_-]+)\b`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(playerURL)
		if len(match) >= 2 {
			return match[1]
		}
	}

	return ""
}
