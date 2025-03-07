package src

import (
	"log"
	"strconv"
)

type BdaBitString struct {
	BasicDataAttribute
	value      []byte
	maxNumBits int
}

func (b *BdaBitString) getMmsDataObj() *Data {
	data := NewData()
	data.bitString = NewBerBitString(nil, b.value, b.maxNumBits)
	return data
}

func (b *BdaBitString) setValueFromMmsDataObj(data *Data) {

	if data.bitString == nil {
		log.Println("ServiceError.TYPE_CONFLICT expected type: bit_string")
		return
		//	throw("ServiceError.TYPE_CONFLICT expected type: bit_string")
	}
	if data.bitString.numBits > b.maxNumBits {
		throw("ServiceError.TYPE_CONFLICT : bit_string is bigger than type's size: " + strconv.Itoa(data.bitString.numBits) + ">" + strconv.Itoa(b.maxNumBits))
	}
	b.value = data.bitString.value
}

func (b *BdaBitString) SetValue(data []byte) {

	b.value = data
}

func NewBdaBitString(objectReference *ObjectReference, fc string, sAddr string, maxNumBits int, dchg bool, dupd bool) *BdaBitString {
	attribute := NewBasicDataAttribute(objectReference, fc, sAddr, dchg, dupd)
	b := &BdaBitString{BasicDataAttribute: *attribute}
	b.maxNumBits = maxNumBits
	b.setDefault()
	return b
}

func (s *BdaBitString) setDefault() {
	s.value = make([]byte, (s.maxNumBits-1)/8+1)
}

func (s *BdaBitString) GetValueString() string {

	return HexStringFromBytes(s.value)
}
