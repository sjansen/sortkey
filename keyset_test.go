package sortkey_test

import (
	"testing"

	"github.com/sjansen/sortkey"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	require := require.New(t)

	ks, err := sortkey.NewKeySet(sortkey.Alpha, sortkey.Base10)
	if err != nil {
		t.Fatalf("unexpected error: %e", err)
	}

	actual, err := ks.Between("", "")
	require.NoError(err)
	require.Equal(sortkey.SortKey("a0"), actual)
}

func TestBetween(t *testing.T) {
	ks, err := sortkey.NewKeySet(sortkey.Alpha, sortkey.Base10)
	if err != nil {
		t.Fatalf("unexpected error: %e", err)
	}

	for i, tc := range []struct {
		a, b     sortkey.SortKey
		expected sortkey.SortKey
	}{
		{"", "", "a0"},
		// Integer Decrement
		{"", "a1", "a0"},
		{"", "a0", "Z9"},
		{"", "b00", "a9"},
		{"", "z00000000000000000000000000", "y9999999999999999999999999"},
		{"", "B0000000000000000000000000", "A99999999999999999999999999"},
		{"", "A000000000000000000000000001", "A00000000000000000000000000"},
		// Integer Increment
		{"a0", "", "a1"},
		{"Z9", "", "a0"},
		{"a9", "", "b00"},
		{"A99999999999999999999999999", "", "B0000000000000000000000000"},
		{"y9999999999999999999999999", "", "z00000000000000000000000000"},
		// Fraction
		{"a0", "a02", "a01"},
		{"a05", "a23", "a1"},
		{"a05", "a06", "a055"},
		{"b00", "b01", "b005"},
	} {
		tc := tc
		t.Run(string('A'+rune(i)), func(t *testing.T) {
			require := require.New(t)

			actual, err := ks.Between(tc.a, tc.b)
			require.NoError(err)
			require.Equal(tc.expected, actual)
		})
	}
}

func TestValidateDigits(t *testing.T) {
	for i, tc := range []struct {
		digits   string
		expected string
	}{
		{"0", "too few digits"},
		{"⠚", "non-ascii digit: ⠚"},
		{"01", ""},
		{"0𐅂", "non-ascii digit: 𐅂"},
		{"aZ", "unsorted digit: Z"},
		{sortkey.Base10, ""},
		{sortkey.Base62, ""},
		{sortkey.Base95, ""},
	} {
		tc := tc
		t.Run(string('A'+rune(i)), func(t *testing.T) {
			_, err := sortkey.NewKeySet(sortkey.Alpha, tc.digits)
			if tc.expected == "" {
				if err != nil {
					t.Fatalf("Unexpected error: %#v", err)
				}
			} else if err == nil {
				t.Fatalf("Expected error: %q", tc.expected)
			}
		})
	}
}

func TestValidateSigils(t *testing.T) {
	for i, tc := range []struct {
		sigils   string
		expected string
	}{
		{"α", "non-ascii sigil: α"},
		{"aΩ", "non-ascii sigil: Ω"},
		{"aZ", "unsorted sigil: Z"},
		{"", ""},
		{sortkey.Alpha, ""},
	} {
		tc := tc
		t.Run(string('A'+rune(i)), func(t *testing.T) {
			_, err := sortkey.NewKeySet(tc.sigils, sortkey.Base10)
			if tc.expected == "" {
				if err != nil {
					t.Fatalf("Unexpected error: %#v", err)
				}
			} else if err == nil {
				t.Fatalf("Expected error: %q", tc.expected)
			}
		})
	}
}
