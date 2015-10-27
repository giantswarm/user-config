package userconfig

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/juju/errgo"
)

var (
	// ByteSizeUnits maps sthg like "kb" to 1000 or "kib" to 1024.
	// See https://en.wikipedia.org/wiki/Kibibyte
	ByteSizeUnits = map[string]uint64{
		"": 1, // Allow empty unit

		"kb":  1000,
		"k":   1000,
		"kib": 1024,

		"m":   1000 * 1000,
		"mb":  1000 * 1000,
		"mib": 1024 * 1024,

		"g":   1000 * 1000 * 1000,
		"gb":  1000 * 1000 * 1000,
		"gib": 1024 * 1024 * 1024,

		"t":   1000 * 1000 * 1000 * 1000,
		"tb":  1000 * 1000 * 1000 * 1000,
		"tib": 1024 * 1024 * 1024 * 1024,
	}
)

type ByteSize string

func (b ByteSize) Valid() bool {
	_, err := b.Bytes()
	return err == nil
}

func (b ByteSize) Bytes() (uint64, error) {
	value, unit, err := b.parse()
	if err != nil {
		return 0, errgo.Mask(err, errgo.Any)
	}

	if factor, ok := ByteSizeUnits[unit]; ok {
		return factor * value, nil
	} else {
		return 0, errgo.Newf("Unknown unit: '%s'", unit)
	}
}

// parse splits b into (uint, unit, error)
func (b ByteSize) parse() (uint64, string, error) {
	r := strings.NewReader(string(b))
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanDigits)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return 0, "", errgo.Mask(err, errgo.Any)
		}
		return 0, "", errgo.Newf("No digits found at begin of input")
	}
	digits := scanner.Text()

	value, err := strconv.ParseUint(digits, 10, 64)
	if err != nil {
		return 0, "", errgo.Mask(err, errgo.Any)
	}

	// Parse the input for the unit
	scanner.Split(bufio.ScanWords)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return 0, "", errgo.Mask(err, errgo.Any)
		}
		return value, "", nil
	}

	unit := scanner.Text()

	// Scan one more time to check for data garbage
	if scanner.Scan() {
		return 0, "", errgo.Newf("Unexpected token: %s", scanner.Text())
	} else {
		if err := scanner.Err(); err != nil {
			return 0, "", errgo.Mask(err, errgo.Any)
		}
	}

	return value, unit, nil
}

func (b ByteSize) String() string {
	return string(b)
}

// ScanDigits implements bufio.SplitFunc for digits (0-9)
func ScanDigits(data []byte, atEOF bool) (advance int, token []byte, err error) {
	fmt.Printf("data=%v\n", data)
	// Skip whitespaces
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			break
		}
	}

	fmt.Printf("1) Start=%v\n", start)

	// Scan digits
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])

		if !unicode.IsDigit(r) {
			fmt.Printf("2) End=%v\n", start+i)
			return i, data[start:i], nil
		}
	}

	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil

}
