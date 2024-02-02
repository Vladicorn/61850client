package src

import (
	"bytes"
	"errors"
	"strconv"
)

type Myexternal2 struct {
	code              []byte
	tag               *BerTag
	directReference   *BerObjectIdentifier
	indirectReference *BerInteger
	encoding          *Encoding
}

func (m *Myexternal2) encode(reverseOS *ReverseByteArrayOutputStream, withTag bool) int {
	if m.code != nil {
		reverseOS.write(m.code)
		if withTag {
			return m.tag.encode(reverseOS) + len(m.code)
		}
		return len(m.code)
	}

	codeLength := 0
	codeLength += m.encoding.encode(reverseOS, nil)

	if m.indirectReference != nil {
		codeLength += m.indirectReference.encode(reverseOS, true)
	}

	if m.directReference != nil {
		codeLength += m.directReference.encode(reverseOS, true)
	}

	codeLength += encodeLength(reverseOS, codeLength)

	if withTag {
		codeLength += m.tag.encode(reverseOS)
	}

	return codeLength
}

func (m *Myexternal2) decode(is *bytes.Buffer, withTag bool) (int, error) {
	var err error
	tlByteCount := 0

	vByteCount := 0

	numDecodedBytes := 0
	berTag := NewEmptyBerTag()

	if withTag {
		tlByteCount += m.tag.decodeAndCheck(is)
	}

	length := NewBerLength()
	tlByteCount += length.decode(is)
	lengthVal := length.val
	vByteCount += berTag.decode(is)

	if berTag.equals(0, 0, 6) {
		m.directReference = NewBerObjectIdentifier(nil)
		vByteCount += m.directReference.decode(is, false)
		vByteCount += berTag.decode(is)
	}

	if berTag.equals(0, 0, 2) {
		m.indirectReference = NewBerInteger(nil, 0)
		vByteCount += m.indirectReference.decode(is, false)
		vByteCount += berTag.decode(is)
	}

	m.encoding = NewEncoding()
	numDecodedBytes, err = m.encoding.decode(is, berTag)
	if err != nil {
		return 0, err
	}
	if numDecodedBytes != 0 {
		vByteCount += numDecodedBytes
		if lengthVal >= 0 && vByteCount == lengthVal {
			return tlByteCount + vByteCount, nil
		}
		vByteCount += berTag.decode(is)
	} else {
		return 0, errors.New("tag does not match mandatory sequence component.")
	}
	if lengthVal < 0 {
		if !berTag.equals(0, 0, 0) {
			return 0, errors.New("Decoded sequence has wrong end of contents octets")
		}
		vByteCount += readEocByte(is)
		return tlByteCount + vByteCount, nil
	}

	return 0, errors.New("Unexpected end of sequence, length tag: " + strconv.Itoa(lengthVal) + ", bytes decoded: " + strconv.Itoa(vByteCount))
}

func NewMyexternal2() *Myexternal2 {
	return &Myexternal2{tag: NewBerTag(0, 32, 8)}
}
