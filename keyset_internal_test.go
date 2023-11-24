package sortkey

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecrementInteger(t *testing.T) {
	ks, err := NewKeySet("!#$%", "abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %e", err)
	}

	for i, tc := range []struct {
		value    SortKey
		expected string
		error    string
	}{{
		value:    "$b",
		expected: "$a",
	}, {
		value:    "$a",
		expected: "#f",
	}, {
		value:    "%aa",
		expected: "$f",
	}, {
		value:    "#a",
		expected: "!ff",
	}, {
		value: "!aa",
		error: "unable to generate smaller value",
	}} {
		tc := tc
		t.Run(string('A'+rune(i)), func(t *testing.T) {
			require := require.New(t)

			parsed, err := ks.parse(tc.value)
			require.NoError(err)

			err = ks.decrementInteger(parsed)
			if tc.error != "" {
				require.EqualError(err, tc.error)
			} else {
				require.NoError(err)

				actual := parsed.String()
				require.Equal(tc.expected, actual)
				require.True(tc.value > SortKey(actual))
			}
		})
	}
}

func TestIncrementInteger(t *testing.T) {
	ks, err := NewKeySet("!#$%", "abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %e", err)
	}

	for i, tc := range []struct {
		value    SortKey
		expected string
		error    string
	}{{
		value:    "$a",
		expected: "$b",
	}, {
		value:    "#f",
		expected: "$a",
	}, {
		value:    "$f",
		expected: "%aa",
	}, {
		value:    "!ff",
		expected: "#a",
	}, {
		value: "%ff",
		error: "unable to generate larger value",
	}} {
		tc := tc
		t.Run(string('A'+rune(i)), func(t *testing.T) {
			require := require.New(t)

			parsed, err := ks.parse(tc.value)
			require.NoError(err)

			err = ks.incrementInteger(parsed)
			if tc.error != "" {
				require.EqualError(err, tc.error)
			} else {
				require.NoError(err)

				actual := parsed.String()
				require.Equal(tc.expected, actual)
				require.True(tc.value < SortKey(actual))
			}
		})
	}
}

func TestNewGenerator(t *testing.T) {
	for i, expected := range []*KeySet{{
		sigils:    "",
		digits:    "01",
		zero:      "0",
		smallest:  []byte{'0'},
		digitsIdx: map[byte]int{'0': 0, '1': 1},
		sigilsIdx: map[byte]int{},
		sigilsLen: map[byte]int{},
	}, {
		sigils:    "#",
		digits:    "01",
		zero:      "#0",
		smallest:  []byte{'#', '0'},
		digitsIdx: map[byte]int{'0': 0, '1': 1},
		sigilsIdx: map[byte]int{'#': 0},
		sigilsLen: map[byte]int{'#': 2},
	}, {
		sigils:    "Za",
		digits:    "01",
		zero:      "a0",
		smallest:  []byte{'Z', '0'},
		digitsIdx: map[byte]int{'0': 0, '1': 1},
		sigilsIdx: map[byte]int{'Z': 0, 'a': 1},
		sigilsLen: map[byte]int{'Z': 2, 'a': 2},
	}, {
		sigils:    "#$%",
		digits:    "abc",
		zero:      "$a",
		smallest:  []byte{'#', 'a'},
		digitsIdx: map[byte]int{'a': 0, 'b': 1, 'c': 2},
		sigilsIdx: map[byte]int{'#': 0, '$': 1, '%': 2},
		sigilsLen: map[byte]int{'#': 2, '$': 2, '%': 3},
	}, {
		sigils:    "ABab",
		digits:    "0123456789",
		zero:      "a0",
		smallest:  []byte{'A', '0', '0'},
		sigilsIdx: map[byte]int{'A': 0, 'B': 1, 'a': 2, 'b': 3},
		sigilsLen: map[byte]int{'A': 3, 'B': 2, 'a': 2, 'b': 3},
		digitsIdx: map[byte]int{
			'0': 0, '1': 1, '2': 2, '3': 3, '4': 4,
			'5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
		},
	}} {
		expected := expected
		t.Run(string('A'+rune(i)), func(t *testing.T) {
			require := require.New(t)

			actual, err := NewKeySet(
				expected.sigils,
				expected.digits,
			)
			require.NoError(err)
			require.Equal(expected, actual)
		})
	}
}

func TestMidpoint(t *testing.T) {
	ks, err := NewKeySet("", Base10)
	require.NoError(t, err)

	for i, tc := range []struct {
		a, b     []byte
		expected []byte
	}{
		{nil, nil, []byte("5")},
		{[]byte("5"), nil, []byte("8")},
		{[]byte("8"), nil, []byte("9")},
		{[]byte("9"), nil, []byte("95")},
		{[]byte("95"), nil, []byte("98")},
		{[]byte("98"), nil, []byte("99")},
		{[]byte("99"), nil, []byte("995")},
		{[]byte("1"), []byte("2"), []byte("15")},
		{[]byte("001"), []byte("001002"), []byte("001001")},
		{[]byte("001"), []byte("001002"), []byte("001001")},
		{nil, []byte("5"), []byte("3")},
		{nil, []byte("3"), []byte("2")},
		{nil, []byte("2"), []byte("1")},
		{nil, []byte("1"), []byte("05")},
		{[]byte("05"), []byte("1"), []byte("08")},
		{nil, []byte("05"), []byte("03")},
		{nil, []byte("03"), []byte("02")},
		{nil, []byte("02"), []byte("01")},
		{nil, []byte("01"), []byte("005")},
		{[]byte("01"), []byte("0111"), []byte("011")},
		{[]byte("499"), []byte("5"), []byte("4995")},
		{[]byte("11"), []byte("1"), nil},
		{[]byte("1"), []byte("1"), nil},
		{[]byte("2"), []byte("1"), nil},
	} {
		tc := tc
		t.Run(string('A'+rune(i)), func(t *testing.T) {
			actual, err := ks.midpoint(tc.a, tc.b)
			if tc.expected == nil {
				require.Error(t, err, string(actual))
			} else if err == nil {
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestParse(t *testing.T) {
	require := require.New(t)

	ks, err := NewKeySet("#$%", "0123456789abcdef")
	require.NoError(err)

	for i, tc := range []struct {
		value    SortKey
		expected string
	}{
		{"", `sortkey too short: ""`},
		{"#", `sortkey too short: "#"`},
		{"+012", `invalid sigil: '+'`},
		{"%aBC", `invalid integer part: "AB"`},
		{"#1dE", `invalid fractional part: "dE"`},
		{"%a", `sortkey too short: "%a"`},
		{"$f0", `trailing zero: "0"`},
		{"%ab", ""},
		{"$0", ""},
		{"#f", ""},
	} {
		tc := tc
		t.Run(string('A'+rune(i)), func(t *testing.T) {
			_, err := ks.parse(tc.value)
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
