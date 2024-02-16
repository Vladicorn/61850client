package src

import (
	"encoding/binary"
	"math"
)

type ClientEventListener struct {
	Data chan EventData
}

type EventData struct {
	Values     map[string]MMSTelegram
	ReportName string
}

func (l *ClientEventListener) associationClosed(err any) {
	close(l.Data)
}

func (l *ClientEventListener) newReport(report *Report) {
	values := make(map[string]MMSTelegram)

	for dataRef, modelNode := range report.reportedDataSetMembersMap {

		if modelNode.getMmsDataObj().structure != nil {
			l.parceModelNode(modelNode.getMmsDataObj(), modelNode.getObjectReference().toString(), dataRef, values)
		} else {
			l.getValue(modelNode.getMmsDataObj(), modelNode.getObjectReference().toString(), dataRef, values)
		}
	}
	data := EventData{
		Values:     values,
		ReportName: report.rptId,
	}
	l.Data <- data
}

func Float32frombytes(bytes []byte) float32 {
	bits := binary.BigEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}
func (l *ClientEventListener) getValue(data *Data, dataName string, dataRef string, values map[string]MMSTelegram) {
	if data.FloatingPoint != nil {
		value := data.FloatingPoint.value[1:]
		mmsTelegram := MMSTelegram{
			Type:  MMSValueType(1),
			Value: Float32frombytes(value),
			Name:  dataName,
		}
		values[dataRef] = mmsTelegram
	}

	if data.integer != nil {
		value := data.integer.value
		mmsTelegram := MMSTelegram{
			Type:  MMSValueType(2),
			Value: value,
		}
		values[dataRef] = mmsTelegram
	}

	if data.utcTime != nil {

		if len(data.utcTime.value) > 0 {
			value := binary.BigEndian.Uint32(data.utcTime.value)
			if mmsTelegram, ok := values[dataRef]; ok {
				mmsTelegram.Time = value
				values[dataRef] = mmsTelegram
			} else {
				mmsTelegram := MMSTelegram{
					Type: MMSValueType(3),
					Time: value,
				}
				values[dataRef] = mmsTelegram
			}
		}
		return
	}

	if data.bool != nil {
		value := data.bool.value
		mmsTelegram := MMSTelegram{
			Type:  MMSValueType(0),
			Value: value,
		}
		values[dataRef] = mmsTelegram
	}

	if data.bitString != nil {
		if mmsTelegram, ok := values[dataRef]; ok {
			mmsTelegram.Quality = data.bitString.value
			values[dataRef] = mmsTelegram
		} else {
			mmsTelegram := MMSTelegram{
				Quality: data.bitString.value,
			}
			values[dataRef] = mmsTelegram
		}
		return
	}
}

func (l *ClientEventListener) parceModelNode(data *Data, dataName string, dataRef string, values map[string]MMSTelegram) {
	structure := data.structure
	for _, modelNode := range structure.seqOf {
		if modelNode.structure != nil {
			l.parceModelNode(modelNode, dataName, dataRef, values)
		} else {
			l.getValue(modelNode, dataName, dataRef, values)
		}
	}

}

func NewClientEventListener() *ClientEventListener {
	data := make(chan EventData)

	return &ClientEventListener{Data: data}
}
