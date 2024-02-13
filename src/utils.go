package src

type MMSValueType uint16

const (
	MMS_VALTYPE_BOOL = MMSValueType(iota)
	MMS_VALTYPE_Float
	MMS_VALTYPE_Int
	MMS_VALTYPE_Time
)

type MMSTelegram struct {
	Type    MMSValueType
	Value   interface{}
	Time    uint32
	Quality []byte
	Name    string
}
