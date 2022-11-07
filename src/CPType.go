package src

import (
	"bytes"
	"strconv"
)

type CPType struct {
	ModeSelector         *ModeSelector
	NormalModeParameters *CPTypeNormalModeParameters
	tag                  *BerTag
	modeSelector         *ModeSelector
	normalModeParameters *NormalModeParameters
}

func (t *CPType) decode(is *bytes.Buffer, withTag bool) int {
	tlByteCount := 0
	vByteCount := 0
	berTag := NewBerTag(0, 0, 0)

	if withTag {
		tlByteCount += t.tag.decodeAndCheck(is)
	}

	length := NewBerLength()
	tlByteCount += length.decode(is)
	lengthVal := length.val

	for vByteCount < lengthVal || lengthVal < 0 {
		vByteCount += berTag.decode(is)
		if berTag.equals(128, 32, 0) {
			t.modeSelector = NewModeSelector()
			vByteCount += t.modeSelector.decode(is, false)
		} else if berTag.equals(128, 32, 2) {
			t.normalModeParameters = NewNormalModeParameters()
			vByteCount += t.normalModeParameters.decode(is, false)
		} else if lengthVal < 0 && berTag.equals(0, 0, 0) {
			vByteCount += readEocByte(is)
			return tlByteCount + vByteCount
		} else {
			throw("tag does not match any set component: " + berTag.toString())
		}
	}
	if vByteCount != lengthVal {
		throw("Length of set does not match length tag, length tag: ", strconv.Itoa(lengthVal), ", actual set length: ", strconv.Itoa(vByteCount))
	}
	return tlByteCount + vByteCount
}

func (t *CPType) encode(reverseOS *ReverseByteArrayOutputStream, withTag bool) int {

	codeLength := 0
	if t.NormalModeParameters != nil {
		codeLength += t.NormalModeParameters.encode(reverseOS, false)
		// writeByte tag: CONTEXT_CLASS, CONSTRUCTED, 2
		reverseOS.writeByte(0xA2)
		codeLength += 1
	}

	codeLength += t.ModeSelector.encode(reverseOS, false)
	// writeByte tag: CONTEXT_CLASS, CONSTRUCTED, 0
	reverseOS.writeByte(0xA0)
	codeLength += 1

	codeLength += encodeLength(reverseOS, codeLength)

	if withTag {
		codeLength += t.tag.encode(reverseOS)
	}

	return codeLength
}

func NewCPType() *CPType {
	return &CPType{tag: NewBerTag(0, 32, 17)}
}
