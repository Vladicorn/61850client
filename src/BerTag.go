package src

import "strconv"

type BerTag struct {
	tagClass  int
	primitive int
	tagNumber int
}

func (t *BerTag) decode(is *ByteBufferInputStream) int {
	nextByte := is.read()
	if nextByte == -1 {
		Throw("Unexpected end of input stream.")
	} else {
		t.tagClass = nextByte & 192
		t.primitive = nextByte & 32
		t.tagNumber = nextByte & 31
		codeLength := 1

		if t.tagNumber == 31 {
			t.tagNumber = 0
			numTagBytes := 0
			//do while
			i := 0
			for i == 0 || (nextByte&128) != 0 {
				i++
				nextByte = is.read()
				if nextByte == -1 {
					Throw("Unexpected end of input stream.")
				}
				codeLength++
				if numTagBytes >= 6 {
					Throw("Tag is too large.")
				}
				t.tagNumber <<= 7
				t.tagNumber |= nextByte & 127
				numTagBytes++
			}
		}
		return codeLength
	}
	return -1
}

func (t *BerTag) equals(identifierClass int, primitive int, tagNumber int) bool {
	return t.tagNumber == tagNumber && t.tagClass == identifierClass && t.primitive == primitive
}

func (t *BerTag) toString() string {
	return "identifier class: " + strconv.Itoa(t.tagClass) + ", primitive: " + strconv.Itoa(t.primitive) + ", tag number: " + strconv.Itoa(t.tagNumber)

}

func NewBerTag() *BerTag {
	return &BerTag{}
}
