package sortkey

import (
	"bytes"
	"strings"
)

type parsedKey struct {
	sigil    byte
	integer  []byte
	fraction []byte
}

func (v *parsedKey) Compare(x *parsedKey) int {
	if v.sigil < x.sigil {
		return -1
	} else if v.sigil > x.sigil {
		return 1
	}
	if tmp := bytes.Compare(v.integer, x.integer); tmp != 0 {
		return tmp
	}
	return bytes.Compare(v.fraction, x.fraction)
}

func (v *parsedKey) integerOnly() *parsedKey {
	return &parsedKey{
		sigil:   v.sigil,
		integer: bytes.Clone(v.integer),
	}
}

func (v *parsedKey) String() string {
	var sb strings.Builder
	sb.Grow(1 + len(v.integer) + len(v.fraction))
	sb.WriteByte(v.sigil)
	sb.Write(v.integer)
	sb.Write(v.fraction)
	return sb.String()
}

func (v *parsedKey) Value() SortKey {
	return SortKey(v.String())
}
