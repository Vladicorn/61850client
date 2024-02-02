package src

import (
	"bytes"
	"errors"
	"strconv"
)

type RLREApdu struct {
	tag             *BerTag
	reason          *ReleaseResponseReason
	userInformation *AssociationInformation
	code            []byte
}

func (a *RLREApdu) decode(is *bytes.Buffer, withTag bool) (int, error) {
	tlByteCount := 0
	vByteCount := 0
	berTag := NewEmptyBerTag()

	if withTag {
		tlByteCount += a.tag.decodeAndCheck(is)
	}

	length := NewBerLength()
	tlByteCount += length.decode(is)
	lengthVal := length.val
	if lengthVal == 0 {
		return tlByteCount, nil
	}
	vByteCount += berTag.decode(is)

	if berTag.equals(128, 0, 0) {
		a.reason = NewReleaseResponseReason()
		vByteCount += a.reason.decode(is, false)
		if lengthVal >= 0 && vByteCount == lengthVal {
			return tlByteCount + vByteCount, nil
		}
		vByteCount += berTag.decode(is)
	}

	if berTag.equals(128, 32, 30) {
		a.userInformation = NewAssociationInformation()
		vByteCountD, err := a.userInformation.decode(is, false)
		if err != nil {
			return 0, err
		}
		vByteCount += vByteCountD
		if lengthVal >= 0 && vByteCount == lengthVal {
			return tlByteCount + vByteCount, nil
		}
		vByteCount += berTag.decode(is)
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
func (a *RLREApdu) encode(reverseOS *ReverseByteArrayOutputStream, withTag bool) int {
	if a.code != nil {
		reverseOS.write(a.code)
		if withTag {
			return a.tag.encode(reverseOS) + len(a.code)
		}
		return len(a.code)
	}

	codeLength := 0
	if a.userInformation != nil {
		codeLength += a.userInformation.encode(reverseOS, false)
		// writeByte tag: CONTEXT_CLASS, CONSTRUCTED, 30
		reverseOS.writeByte(0xBE)
		codeLength += 1
	}

	if a.reason != nil {
		codeLength += a.reason.encode(reverseOS, false)
		// writeByte tag: CONTEXT_CLASS, PRIMITIVE, 0
		reverseOS.writeByte(0x80)
		codeLength += 1
	}

	codeLength += encodeLength(reverseOS, codeLength)

	if withTag {
		codeLength += a.tag.encode(reverseOS)
	}

	return codeLength
}

func NewRLREApdu() *RLREApdu {
	return &RLREApdu{}
}
