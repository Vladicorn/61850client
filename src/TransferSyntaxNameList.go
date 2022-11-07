package src

import (
	"bytes"
	"strconv"
)

type TransferSyntaxNameList struct {
	tag   *BerTag
	seqOf []*TransferSyntaxName
	code  []byte
}

func (l *TransferSyntaxNameList) encode(reverseOS *ReverseByteArrayOutputStream, withTag bool) int {
	if l.code != nil {
		reverseOS.writeByte(l.code)
		if withTag {
			return l.tag.encode(reverseOS) + len(l.code)
		}
		return len(l.code)
	}

	codeLength := 0
	for i := len(l.seqOf) - 1; i >= 0; i-- {
		codeLength += l.seqOf[i].encode(reverseOS, true)
	}

	codeLength += encodeLength(reverseOS, codeLength)

	if withTag {
		codeLength += l.tag.encode(reverseOS)
	}

	return codeLength
}
func (l *TransferSyntaxNameList) decode(is *bytes.Buffer, withTag bool) int {

	tlByteCount := 0
	vByteCount := 0
	berTag := NewBerTag(0, 0, 0)
	if withTag {
		tlByteCount += l.tag.decodeAndCheck(is)
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

		if !berTag.equals(0, 0, 6) {
			throw("tag does not match mandatory sequence of/set of component.")
		}
		element := NewTransferSyntaxName()
		vByteCount += element.decode(is, false)
		l.seqOf = append(l.seqOf, element)
	}
	if lengthVal >= 0 && vByteCount != lengthVal {
		throw("Decoded SequenceOf or SetOf has wrong length. Expected ", strconv.Itoa(lengthVal), " but has ", strconv.Itoa(vByteCount))
	}
	return tlByteCount + vByteCount
}

func NewTransferSyntaxNameList() *TransferSyntaxNameList {
	return &TransferSyntaxNameList{tag: NewBerTag(0, 32, 16)}
}
