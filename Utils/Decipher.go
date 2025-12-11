package Utils

import (
	"crypto/sha1"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (

	// JavaScript Regexes for Parsing

	JSVarStr         = `[a-zA-Z_\$]\w*`
	JSSingleQuoteStr = `'[^'\\]*(:?\\[\s\S][^'\\]*)*'`
	JSDoubleQuoteStr = `"[^"\\]*(:?\\[\s\S][^"\\]*)*"`
	JSQuoteStr       = fmt.Sprintf(`(?:%s|%s)`, JSSingleQuoteStr, JSDoubleQuoteStr)

	JSKeyStr         = fmt.Sprintf(`(?:%s|%s)`, JSVarStr, JSQuoteStr)
	JSPropStr        = fmt.Sprintf(`(?:\.%s|\[%s\])`, JSVarStr, JSQuoteStr)
	JSEmptyStr       = `(?:''|"")`

	ReverseStr = `:function\(a\)\{(?:return )?a\.reverse\(\)\}`
	SliceStr   = `:function\(a,b\)\{return a\.slice\(b\)\}`
	SpliceStr  = `:function\(a,b\)\{a\.splice\(0,b\)\}`
	SwapStr    = `:function\(a,b\)\{var c=a\[0\];a\[0\]=a\[b(?:%a\.length)?\];a\[b(?:%a\.length)?\]=c(?:;return a)?\}`

	ActionsObjRegex  = regexp.MustCompile(fmt.Sprintf(`var (%s)=\{((?:(?:%s%s|%s%s|%s%s|%s%s),?\r?\n?)+)\};`, JSVarStr, JSKeyStr, ReverseStr, JSKeyStr, SliceStr, JSKeyStr, SpliceStr, JSKeyStr, SwapStr))
	ActionsFuncRegex = regexp.MustCompile(fmt.Sprintf(`function(?: %s)?\(a\)\{a=a\.split\(%s\);\s*((?:(?:a=)?%s%s\(a,\d+\);)+)return a\.join\(%s\)\}`, JSVarStr, JSEmptyStr, JSVarStr, JSPropStr, JSEmptyStr))

	ReverseRegex = regexp.MustCompile(fmt.Sprintf(`(?:^|,)(%s)%s`, JSKeyStr, ReverseStr))
	SliceRegex   = regexp.MustCompile(fmt.Sprintf(`(?:^|,)(%s)%s`, JSKeyStr, SliceStr))
	SpliceRegex  = regexp.MustCompile(fmt.Sprintf(`(?:^|,)(%s)%s`, JSKeyStr, SpliceStr))

	SwapRegex    = regexp.MustCompile(fmt.Sprintf(`(?:^|,)(%s)%s`, JSKeyStr, SwapStr))
)

func GenerateHashFromCookies(Cookies string, Origin string) (string, error) {

	sidMatch := regexp.MustCompile(`SAPISID=([^;]+)`).FindStringSubmatch(Cookies)

	if len(sidMatch) < 2 {

		return "", fmt.Errorf("SAPISID not found in cookies")

	}

	SAPIID := sidMatch[1]

	Timestamp := time.Now().Unix()
	Input := fmt.Sprintf("%d %s %s", Timestamp, SAPIID, Origin)
	Hash := fmt.Sprintf("%x", sha1.Sum([]byte(Input)))

	return fmt.Sprintf("SAPISIDHASH %d_%s", Timestamp, Hash), nil

}

func Decipher(Tokens []string, Sig string) string {

	arr := strings.Split(Sig, "")

	for _, Token := range Tokens {
		if len(Token) < 2 {
			continue
		}

		Pos, _ := strconv.Atoi(Token[1:])

		switch Token[0] {

		case 'r':

			// Reverse

			for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {

				arr[i], arr[j] = arr[j], arr[i]

			}

		case 's':

			// Slice

			if Pos < len(arr) {

				arr = arr[Pos:]

			}

		case 'p':

			// Splice

			if Pos < len(arr) {

				arr = arr[Pos:]

			}

		case 'w':

			// Swap

			swapPos := Pos % len(arr)
			arr[0], arr[swapPos] = arr[swapPos], arr[0]

		}

	}

	return strings.Join(arr, "")

}

func ExtractTokens(Body string) []string {

	Objects := ActionsObjRegex.FindStringSubmatch(Body)
	Functions := ActionsFuncRegex.FindStringSubmatch(Body)

	if len(Objects) < 3 || len(Functions) < 2 {
		return nil
	}

	Obj := strings.ReplaceAll(Objects[1], "$", "\\$")
	ObjBody := strings.ReplaceAll(Objects[2], "$", "\\$")
	FuncBody := strings.ReplaceAll(Functions[1], "$", "\\$")

	reverseKeys := extractKey(ReverseRegex, ObjBody)
	sliceKeys := extractKey(SliceRegex, ObjBody)
	spliceKeys := extractKey(SpliceRegex, ObjBody)
	swapKeys := extractKey(SwapRegex, ObjBody)

	Keys := fmt.Sprintf("(%s|%s|%s|%s)", reverseKeys, sliceKeys, spliceKeys, swapKeys)
	tokenizeRegex := regexp.MustCompile(fmt.Sprintf(`(?:a=)?%s(?:\.%s|\[(?:'%s'|"%s")\])\(a,(\d+)\)`, Obj, Keys, Keys, Keys))

	var Tokens []string

	matches := tokenizeRegex.FindAllStringSubmatch(FuncBody, -1)

	for _, Result := range matches {

		if len(Result) < 5 {
			
			continue

		}

		key := Result[1]

		if key == "" {

			key = Result[2]

		}

		if key == "" {

			key = Result[3]

		}

		switch key {

		case reverseKeys:

			Tokens = append(Tokens, "r")

		case sliceKeys:

			Tokens = append(Tokens, "s"+Result[4])

		case spliceKeys:

			Tokens = append(Tokens, "p"+Result[4])

		case swapKeys:

			Tokens = append(Tokens, "w"+Result[4])

		}

	}

	return Tokens

}

func extractKey(regex *regexp.Regexp, body string) string {

	match := regex.FindStringSubmatch(body)

	if len(match) < 2 {

		return ""

	}

	key := strings.ReplaceAll(match[1], "$", "\\$")
	key = strings.Trim(key, "'\"")

	return key

}