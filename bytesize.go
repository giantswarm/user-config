package userconfig

import (
	"bufio"
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

	UnknownByteSizeUnitError                  = errgo.Newf("Unknown unit provided")
	InvalidByteSizeFormatNoDigitsError        = errgo.Newf("No digits found at beginning of input")
	InvalidByteSizeFormatUnexpectedTokenError = errgo.Newf("Unexpected token")
)

func IsUnknownByteSizeUnit(err error) bool {
	return errgo.Cause(err) == UnknownByteSizeUnitError
}

func IsInvalidByteSizeFormatNoDigits(err error) bool {
	return errgo.Cause(err) == InvalidByteSizeFormatNoDigitsError
}

func IsInvalidByteSizeFormatUnexpectedToken(err error) bool {
	return errgo.Cause(err) == InvalidByteSizeFormatUnexpectedTokenError
}

type ByteSize string

func (b ByteSize) Equals(other ByteSize) bool {
	myBytes, err := b.Bytes()
	if err != nil {
		return false
	}

	otherBytes, err := other.Bytes()
	if err != nil {
		return false
	}

	return myBytes == otherBytes
}

// Valid returns a bool indicating whether this ByteSize value can successfully be parsed.
func (b ByteSize) Valid() bool {
	_, err := b.Bytes()
	return err == nil
}

// Bytes parses the number of bytes describes by this ByteSize.
// If the value of b is unparsable, an error is returned. On success, the value in bytes is returend.
// Example: if the value contains "3 kb", 3000 will be returned.
func (b ByteSize) Bytes() (uint64, error) {
	value, unit, err := b.Parse()
	if err != nil {
		return 0, errgo.Mask(err, errgo.Any)
	}

	if factor, ok := ByteSizeUnits[unit]; ok {
		return factor * value, nil
	} else {
		return 0, errgo.WithCausef(err, UnknownByteSizeUnitError, unit)
	}
}

// parse splits b into (uint, unit, error)
func (b ByteSize) Parse() (uint64, string, error) {
	r := strings.NewReader(string(b))
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanDigits)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return 0, "", errgo.Mask(err, errgo.Any)
		}
		return 0, "", errgo.WithCausef(nil, InvalidByteSizeFormatNoDigitsError, "Input: %s", string(b))
	}
	digits := scanner.Text()

	if digits == "" {
		return 0, "", errgo.WithCausef(nil, InvalidByteSizeFormatNoDigitsError, "Input: %s", string(b))
	}

	value, err := strconv.ParseUint(digits, 10, 64)
	if err != nil {
		// This should never happen, as the ScanDigits() function shouldn't have returned that token
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
		return 0, "", errgo.WithCausef(nil, InvalidByteSizeFormatUnexpectedTokenError, scanner.Text())
	} else {
		if err := scanner.Err(); err != nil {
			return 0, "", errgo.Mask(err, errgo.Any)
		}
	}

	return value, unit, nil
}

// String returns the string value of this ByteSize as initially provided.
func (b ByteSize) String() string {
	return string(b)
}

// ScanDigits implements bufio.SplitFunc for digits (0-9). Leading whitespaces (unicode.IsSpace) are ignored.
// May return an empty token, if no digits are found at the beginning.
func ScanDigits(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip whitespaces
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			break
		}
	}

	// Scan digits
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])

		if !unicode.IsDigit(r) {
			return i, data[start:i], nil
		}
	}

	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil
}
