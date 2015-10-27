package userconfig

import (
	"testing"
)

func TestByteSizeParse(t *testing.T) {
	tests := []struct {
		input string
		valid bool
		value uint64
	}{
		{" 1 ", true, 1},
		{"3", true, 3},
		{"1", true, 1},
		{"1000", true, 1000},
		{"1024", true, 1024},
		{"1 kb", true, 1000},
		{"1 kib", true, 1024},
		{"2kb", true, 2000},
		{"1kb man", false, 0},
		{"1k", true, 1000},
		{"1kb ", true, 1000},
		{"1kb man", false, 0},
		{" 1 kib", true, 1024},
		{"kb", false, 0},
		{"1e10 gb", false, 0},
		{"1 egg", false, 0},
	}

	for idx, test := range tests {
		b := ByteSize(test.input)

		v, err := b.Bytes()
		if test.valid {
			if err != nil {
				t.Errorf("Test %d, Bytes(%s) returned unexpected error: %v", idx, test.input, err)
			}
		} else {
			if err == nil {
				t.Errorf("Test %d, Bytes(%s) expected error, but received nil.", idx, test.input)
			}
		}
		if v != test.value {
			t.Errorf("Test %d, Bytes(%s): Expected %v, got %v\n", idx, test.input, test.value, v)
		}
	}
}

func TestByteSize__UnknownByteSizeUnitError(t *testing.T) {
	tests := []struct {
		input   string
		checker func(err error) bool
	}{
		{"1 egg", IsUnknownByteSizeUnit},
		{"dog", IsInvalidByteSizeFormatNoDigits},
		{"dog 2", IsInvalidByteSizeFormatNoDigits},
		{"100 little elefants", IsInvalidByteSizeFormatUnexpectedToken},
	}

	for idx, test := range tests {
		_, err := ByteSize(test.input).Bytes()
		if err == nil {
			t.Errorf("Test %d: Expected error for input '%s'", idx, test.input)
		} else if !test.checker(err) {
			t.Errorf("Test %d: Expected checker to return true, for input '%s'. Error is '%v'", idx, test.input, err)
		}
	}
}

func TestByteSize__Equals(t *testing.T) {
	b1 := ByteSize("1 mb")
	b2 := ByteSize("1m")

	if !b1.Equals(b2) {
		t.Fatalf("Expected '%s' and '%s' to be equal.", b1, b2)
	}
}
