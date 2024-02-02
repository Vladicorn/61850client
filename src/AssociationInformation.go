package src

import (
	"bytes"
	"errors"
	"strconv"
)

type AssociationInformation struct {
	seqOf []*Myexternal
	tag   *BerTag
	code  []byte
}

func (a *AssociationInformation) encode(reverseOS *ReverseByteArrayOutputStream, withTag bool) int {
	if a.code != nil {
		reverseOS.write(a.code)
		if withTag {
			return a.tag.encode(reverseOS) + len(a.code)
		}
		return len(a.code)
	}

	codeLength := 0
	for i := len(a.seqOf) - 1; i >= 0; i-- {
		codeLength += a.seqOf[i].encode(reverseOS, true)
	}

	codeLength += encodeLength(reverseOS, codeLength)

	if withTag {
		codeLength += a.tag.encode(reverseOS)
	}

	return codeLength
}

func (a *AssociationInformation) decode(is *bytes.Buffer, withTag bool) (int, error) {
	tlByteCount := 0
	vByteCount := 0
	berTag := NewEmptyBerTag()
	if withTag {
		tlByteCount += a.tag.decodeAndCheck(is)
	}

	length := NewBerLength()
	tlByteCount += length.decode(is)
	lengthVal := length.val

	for vByteCount < lengthVal || lengthVal < 0 {
		vByteCount += berTag.decode(is)

		if lengthVal < 0 && berTag.equals(0, 0, 0) {
			vByteCount += readEocByte(is)
			break
		}

		if !berTag.equals(0, 32, 8) {
			return 0, errors.New("tag does not match mandatory sequence of/set of component.")
		}
		element := NewMyexternal()
		vByteCountD, err := element.decode(is, false)
		if err != nil {
			return 0, err
		}
		vByteCount += vByteCountD
		a.seqOf = append(a.seqOf, element)
	}
	if lengthVal >= 0 && vByteCount != lengthVal {
		return 0, errors.New("Decoded SequenceOf or SetOf has wrong length. Expected " + strconv.Itoa(lengthVal) + " but has " + strconv.Itoa(vByteCount))
	}
	return tlByteCount + vByteCount, nil
}

func NewAssociationInformation() *AssociationInformation {
	return &AssociationInformation{tag: NewBerTag(0, 32, 16)}
}
