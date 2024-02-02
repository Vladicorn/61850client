package src

import (
	"bytes"
	"errors"
)

type ACSEApdu struct {
	aarq *AARQApdu
	aare *AAREApdu
	rlrq *RLRQApdu
	rlre *RLREApdu
}

func (a *ACSEApdu) encode(reverseOS *ReverseByteArrayOutputStream) int {
	codeLength := 0
	if a.aarq != nil {
		codeLength += a.aarq.encode(reverseOS, true)
		return codeLength
	}

	throw("Error encoding WriteResponseCHOICE: No element of WriteResponseCHOICE was selected.")
	return -1
}

func (a *ACSEApdu) decode(is *bytes.Buffer, berTag *BerTag) (int, error) {

	tlvByteCount := 0
	tagWasPassed := berTag != nil

	if berTag == nil {
		berTag = NewEmptyBerTag()
		tlvByteCount += berTag.decode(is)
	}

	if berTag.equals(64, 32, 0) {
		a.aarq = NewAARQApdu()
		tlvByteCountD, err := a.aarq.decode(is, false)
		if err != nil {
			return 0, err
		}
		tlvByteCount += tlvByteCountD
		return tlvByteCount, nil
	}

	if berTag.equals(64, 32, 1) {
		a.aare = NewAAREApdu()
		tlvByteCountD, err := a.aare.decode(is, false)
		if err != nil {
			return 0, err
		}
		tlvByteCount += tlvByteCountD
		return tlvByteCount, nil
	}

	if berTag.equals(64, 32, 2) {
		a.rlrq = NewRLRQApdu()
		tlvByteCountD, err := a.rlrq.decode(is, false)
		if err != nil {
			return 0, err
		}
		tlvByteCount += tlvByteCountD
		return tlvByteCount, nil
	}

	if berTag.equals(64, 32, 3) {
		a.rlre = NewRLREApdu()
		tlvByteCountD, err := a.rlre.decode(is, false)
		if err != nil {
			return 0, err
		}
		tlvByteCount += tlvByteCountD
		return tlvByteCount, nil
	}

	if tagWasPassed {
		return 0, nil
	}

	return 0, errors.New("Error decoding WriteResponseCHOICE: tag " + berTag.toString() + " matched to no item.")
}

func NewACSEApdu() *ACSEApdu {
	return &ACSEApdu{}
}
