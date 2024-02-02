package src

import (
	"bytes"
	"errors"
	"strconv"
)

type FullyEncodedData struct {
	seqOf []*PDVList
	tag   *BerTag
	code  []byte
}

func (d *FullyEncodedData) getPDVList() []*PDVList {
	if d.seqOf == nil {
		d.seqOf = make([]*PDVList, 0)
	}
	return d.seqOf

}

func (d *FullyEncodedData) encode(reverseOS *ReverseByteArrayOutputStream, withTag bool) int {
	if d.code != nil {
		reverseOS.write(d.code)
		if withTag {
			return d.tag.encode(reverseOS) + len(d.code)
		}
		return len(d.code)
	}

	codeLength := 0
	for i := len(d.seqOf) - 1; i >= 0; i-- {
		codeLength += d.seqOf[i].encode(reverseOS, true)
	}

	codeLength += encodeLength(reverseOS, codeLength)

	if withTag {
		codeLength += d.tag.encode(reverseOS)
	}

	return codeLength
}

func (d *FullyEncodedData) decode(is *bytes.Buffer, withTag bool) (int, error) {
	tlByteCount := 0
	vByteCount := 0
	berTag := NewEmptyBerTag()
	if withTag {
		tlByteCount += d.tag.decodeAndCheck(is)
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

		if !berTag.equals(0, 32, 16) {
			errors.New("tag does not match mandatory sequence of/set of component")
		}
		element := NewPDVList()
		vByteCountD, err := element.decode(is, false)
		if err != nil {
			return 0, err
		}
		vByteCount += vByteCountD
		d.seqOf = append(d.seqOf, element)
	}
	if lengthVal >= 0 && vByteCount != lengthVal {

		errors.New("Decoded SequenceOf or SetOf has wrong length. Expected " + strconv.Itoa(lengthVal) + " but has " + strconv.Itoa(vByteCount))
	}
	return tlByteCount + vByteCount, nil
}

func NewFullyEncodedData() *FullyEncodedData {
	return &FullyEncodedData{tag: NewBerTag(0, 32, 16), seqOf: make([]*PDVList, 0)}
}
