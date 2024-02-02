package src

import (
	"bytes"
	"errors"
)

type MyexternalEncoding struct {
	singleASN1Type *BerAny
	code           []byte
	octetAligned   *BerOctetString
	arbitrary      *BerBitString
}

func (e *MyexternalEncoding) encode(reverseOS *ReverseByteArrayOutputStream) int {
	if e.code != nil {
		reverseOS.write(e.code)
		return len(e.code)
	}

	codeLength := 0
	sublength := 0

	if e.arbitrary != nil {
		codeLength += e.arbitrary.encode(reverseOS, false)
		// writeByte tag: CONTEXT_CLASS, PRIMITIVE, 2
		reverseOS.writeByte(0x82)
		codeLength += 1
		return codeLength
	}

	if e.octetAligned != nil {
		codeLength += e.octetAligned.encode(reverseOS, false)
		// writeByte tag: CONTEXT_CLASS, PRIMITIVE, 1
		reverseOS.writeByte(0x81)
		codeLength += 1
		return codeLength
	}

	if e.singleASN1Type != nil {
		sublength = e.singleASN1Type.encode(reverseOS)
		codeLength += sublength
		codeLength += encodeLength(reverseOS, sublength)
		// writeByte tag: CONTEXT_CLASS, CONSTRUCTED, 0
		reverseOS.writeByte(0xA0)
		codeLength += 1
		return codeLength
	}

	throw("Error encoding WriteResponseCHOICE: No element of WriteResponseCHOICE was selected.")
	return 0
}

func (e *MyexternalEncoding) decode(is *bytes.Buffer, berTag *BerTag) (int, error) {

	tlvByteCount := 0
	tagWasPassed := berTag != nil

	if berTag == nil {
		berTag = NewEmptyBerTag()
		tlvByteCount += berTag.decode(is)
	}

	if berTag.equals(128, 32, 0) {

		length := NewBerLength()
		tlvByteCount += length.decode(is)
		e.singleASN1Type = NewBerAny(nil)
		tlvByteCountD, err := e.singleASN1Type.decode(is, nil)
		if err != nil {
			return 0, err
		}
		tlvByteCount += tlvByteCountD
		tlvByteCount += length.readEocIfIndefinite(is)
		return tlvByteCount, nil
	}

	if berTag.equals(128, 0, 1) {
		e.octetAligned = NewBerOctetString(nil)
		tlvByteCount += e.octetAligned.decode(is, false)
		return tlvByteCount, nil
	}

	if berTag.equals(128, 0, 2) {
		e.arbitrary = NewBerBitString(nil, nil, 0)
		tlvByteCount += e.arbitrary.decode(is, false)
		return tlvByteCount, nil
	}

	if tagWasPassed {
		return 0, nil
	}

	return 0, errors.New("Error decoding WriteResponseCHOICE: tag " + berTag.toString() + " matched to no item.")
}
func NewMyexternalEncoding() *MyexternalEncoding {
	return &MyexternalEncoding{}
}
