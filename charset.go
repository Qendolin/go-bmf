package bmf

import (
	"fmt"
	"strconv"
	"strings"
)

// CharsetTable enumerates known charset values
// source https://docs.microsoft.com/en-us/previous-versions/windows/desktop/bb322881(v=vs.85)
var CharsetTable = map[int]string{
	186: "Baltic",
	77:  "Mac",
	204: "Russian",
	238: "EastEurope",
	222: "Thai",
	163: "Vietnamese",
	162: "Turkish",
	161: "Greek",
	178: "Arabic",
	177: "Hebrew",
	130: "Johab",
	255: "Oem",
	136: "ChineseBig5",
	134: "GB2312",
	129: "Hangul",
	128: "ShiftJIS",
	2:   "Symbol",
	1:   "Default",
	0:   "Ansi",
}

// LookupCharset gets the name of a charset or value as a decimal string when not found
func LookupCharset(charsetEnum int) (name string, found bool) {
	if charset, ok := CharsetTable[charsetEnum]; ok {
		return charset, true
	}
	return fmt.Sprintf("%d", charsetEnum), false
}

// LookupCharsetValue gets the enum value of a charset or converts a a decimal string to a number
func LookupCharsetValue(charset string) (value int, found bool) {
	charset = strings.ToLower(charset)

	for k, v := range CharsetTable {
		if strings.ToLower(v) == charset {
			return k, true
		}
	}

	value, err := strconv.Atoi(charset)
	return value, err == nil
}
