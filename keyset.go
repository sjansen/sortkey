package sortkey

import (
	"bytes"
	"errors"
	"fmt"
	"math"
)

const Alpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const NoVowels = "BCDFGHJKLMNPQRSTVWXZbcdfghjklmnpqrstvwxz"

const Base10 = "0123456789"
const Base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const Base95 = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

type KeySet struct {
	sigils   string
	digits   string
	zero     SortKey
	smallest []byte

	sigilsIdx map[byte]int
	sigilsLen map[byte]int
	digitsIdx map[byte]int
}

type SortKey string

func NewKeySet(sigils, digits string) (*KeySet, error) {
	if err := validateSigils(sigils); err != nil {
		return nil, err
	}
	if err := validateDigits(digits); err != nil {
		return nil, err
	}

	ks := &KeySet{
		digits:    digits,
		sigils:    sigils,
		digitsIdx: make(map[byte]int, len(digits)),
		sigilsIdx: make(map[byte]int, len(sigils)),
		sigilsLen: make(map[byte]int, len(sigils)),
	}

	// digitsIdx
	for i, digit := range digits {
		ks.digitsIdx[byte(digit)] = i
	}

	// sigilsIdx
	// sigilsLen
	midpoint := len(sigils) / 2
	for i, sigil := range sigils {
		ks.sigilsIdx[byte(sigil)] = i
		if i < midpoint {
			ks.sigilsLen[byte(sigil)] = midpoint - i + 1
		} else {
			ks.sigilsLen[byte(sigil)] = 2 + i - midpoint
		}
	}

	// smallest
	if len(sigils) < 1 {
		ks.smallest = []byte{digits[0]}
	} else {
		n := ks.sigilsLen[sigils[0]]
		var bb bytes.Buffer
		bb.Grow(n)
		bb.WriteByte(sigils[0])
		for i := 1; i < n; i++ {
			bb.WriteByte(digits[0])
		}
		ks.smallest = bb.Bytes()
	}

	// zero
	if len(sigils) < 1 {
		ks.zero = SortKey(digits[0])
	} else {
		ks.zero = SortKey(string(sigils[midpoint]) + string(digits[0]))
	}

	return ks, nil
}

func validateDigits(digits string) error {
	if len(digits) < 2 {
		return &ConfigError{"too few digits"}
	}

	runes := []rune(digits)
	prev := runes[0]
	if prev >= 128 {
		return &ConfigError{
			fmt.Sprintf("non-ascii digit: %#v", prev),
		}
	}

	for i := 1; i < len(runes); i++ {
		next := runes[i]
		if next >= 128 {
			return &ConfigError{
				fmt.Sprintf("non-ascii digit: %#v", next),
			}
		}

		if prev >= next {
			return &ConfigError{
				fmt.Sprintf("unsorted digit: %#v", next),
			}
		}
		prev = next
	}

	return nil
}

func validateSigils(sigils string) error {
	if len(sigils) < 1 {
		return nil
	}

	runes := []rune(sigils)
	prev := runes[0]
	if prev >= 128 {
		return &ConfigError{
			fmt.Sprintf("non-ascii sigil: %#v", prev),
		}
	}

	for i := 1; i < len(runes); i++ {
		next := runes[i]
		if next >= 128 {
			return &ConfigError{
				fmt.Sprintf("non-ascii sigil: %#v", next),
			}
		}

		if prev >= next {
			return &ConfigError{
				fmt.Sprintf("unsorted sigil: %#v", next),
			}
		}
		prev = next
	}

	return nil
}

func (ks *KeySet) Between(a, b SortKey) (SortKey, error) {
	switch {
	case a == "":
		if b == "" {
			return ks.zero, nil
		}

		pb, err := ks.parse(b)
		if err != nil {
			return "", err
		}

		switch {
		case bytes.Equal(pb.integer, ks.smallest):
			// TODO return ErrUnderflow instead?
			if fb, err := ks.midpoint(nil, pb.fraction); err != nil {
				return "", err
			} else {
				pb.fraction = fb
			}
		case len(pb.fraction) > 0:
			pb.fraction = nil
		default:
			if err = ks.decrementInteger(pb); err != nil {
				return "", err
			}
		}
		return pb.Value(), nil
	case b == "":
		pa, err := ks.parse(a)
		if err != nil {
			return "", err
		}

		// TODO handle ErrOverflow?
		if err = ks.incrementInteger(pa); err != nil {
			return "", err
		}
		return pa.Value(), nil
	}

	pa, err := ks.parse(a)
	if err != nil {
		return "", err
	}

	pb, err := ks.parse(b)
	if err != nil {
		return "", err
	}

	if pa.sigil == pb.sigil && bytes.Equal(pa.integer, pb.integer) {
		if fa, err := ks.midpoint(pa.fraction, pb.fraction); err != nil {
			return "", err
		} else {
			pa.fraction = fa
		}
		return pa.Value(), nil
	}

	ai := pa.integerOnly()
	if err = ks.incrementInteger(ai); err != nil {
		return "", err
	}
	if ai.Compare(pb) < 0 {
		return ai.Value(), nil
	}

	tmp, err := ks.midpoint(pa.fraction, nil)
	if err != nil {
		return "", nil
	}
	pa.fraction = tmp
	return pa.Value(), nil
}

func (ks *KeySet) decrementInteger(v *parsedKey) error {
	borrow := true
	for i := len(v.integer) - 1; borrow && i >= 0; i-- {
		idx := ks.digitsIdx[v.integer[i]] - 1
		if idx == -1 {
			v.integer[i] = ks.digits[len(ks.digits)-1]
		} else {
			v.integer[i] = ks.digits[idx]
			borrow = false
		}
	}
	if borrow {
		idx := ks.sigilsIdx[v.sigil] - 1
		if idx == -1 {
			return ErrUnderflow
		}

		n := len(v.integer) + 1
		v.sigil = ks.sigils[idx]
		switch {
		case n < ks.sigilsLen[v.sigil]:
			v.integer = append(v.integer, ks.digits[len(ks.digits)-1])
		case n > ks.sigilsLen[v.sigil]:
			v.integer = v.integer[1:]
		}
	}
	return nil
}

func (ks *KeySet) incrementInteger(v *parsedKey) error {
	carry := true
	for i := len(v.integer) - 1; carry && i >= 0; i-- {
		idx := ks.digitsIdx[v.integer[i]] + 1
		if idx == len(ks.digits) {
			v.integer[i] = ks.digits[0]
		} else {
			v.integer[i] = ks.digits[idx]
			carry = false
		}
	}
	if carry {
		idx := ks.sigilsIdx[v.sigil] + 1
		if idx == len(ks.sigils) {
			return ErrOverflow
		}

		n := len(v.integer) + 1
		v.sigil = ks.sigils[idx]
		switch {
		case n < ks.sigilsLen[v.sigil]:
			v.integer = append(v.integer, ks.digits[0])
		case n > ks.sigilsLen[v.sigil]:
			v.integer = v.integer[1:]
		}
	}
	return nil
}

func (ks *KeySet) midpoint(a, b []byte) ([]byte, error) {
	if len(b) > 0 && bytes.Compare(a, b) >= 0 {
		return nil, errors.New("a >= b") // TODO custom error
	}
	prefix := ks.midpointPrefix(a, b)
	if n := len(prefix); n > 0 {
		if n < len(a) {
			a = a[n:]
		} else {
			a = nil
		}
		b = b[n:]
	}
	suffix := ks.midpointSuffix(a, b)
	result := bytes.Join([][]byte{prefix, suffix}, nil)
	return result, nil
}

func (ks *KeySet) midpointPrefix(a, b []byte) []byte {
	zero := ks.digits[0]
	i := 0
	for ; i < len(b); i++ {
		tmp := zero
		if i < len(a) {
			tmp = a[i]
		}
		if tmp != b[i] {
			break
		}
	}
	if i > 0 {
		return b[0:i]
	}
	return nil
}

func (ks *KeySet) midpointSuffix(a, b []byte) []byte {
	result := make([]byte, 0, len(a)+1)
	for {
		digitA := 0
		if len(a) > 0 {
			digitA = ks.digitsIdx[a[0]]
		}
		digitB := len(ks.digits)
		if len(b) > 0 {
			digitB = ks.digitsIdx[b[0]]
		}
		if digitB-digitA > 1 {
			midDigit := int(math.Round(0.5 * float64(digitA+digitB)))
			return append(result, ks.digits[midDigit])
		}

		if len(b) > 1 {
			return append(result, b[0])
		}

		result = append(result, ks.digits[digitA])
		if len(a) > 0 {
			a = a[1:]
		}
		b = nil
	}
}

func (ks *KeySet) parse(value SortKey) (*parsedKey, error) {
	if value == "" {
		return nil, &InvalidValueError{`sortkey too short: ""`}
	}

	sigil := value[0]
	n, ok := ks.sigilsLen[sigil]
	if !ok {
		return nil, &InvalidValueError{
			fmt.Sprintf("invalid sigil: %c", sigil),
		}
	}
	if len(value) < n {
		return nil, &InvalidValueError{
			fmt.Sprintf("sortkey too short: %q", value),
		}
	}

	tmp := string(value)
	result := &parsedKey{
		sigil:    sigil,
		integer:  []byte(tmp[1:n]),
		fraction: []byte(tmp[n:]),
	}
	if err := ks.validateParsed(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (ks *KeySet) validateParsed(parsed *parsedKey) error {
	for _, x := range parsed.integer {
		if _, ok := ks.digitsIdx[x]; !ok {
			return &InvalidValueError{
				fmt.Sprintf("invalid integer part: %q", parsed.integer),
			}
		}
	}
	for _, x := range parsed.fraction {
		if _, ok := ks.digitsIdx[x]; !ok {
			return &InvalidValueError{
				fmt.Sprintf("invalid fractional part: %q", parsed.fraction),
			}
		}
	}
	n := len(parsed.fraction)
	if n > 0 && parsed.fraction[n-1] == ks.digits[0] {
		return &InvalidValueError{
			fmt.Sprintf("trailing zero: %q", ks.digits[0]),
		}
	}
	return nil
}
