package src

import (
	"strconv"
)

type FcDataObjectI interface {
	FcModelNodeI
	GetObjectReference() *ObjectReference
}

type FcDataObject struct {
	FcModelNode
}

func (n *FcDataObject) GetObjectReference() *ObjectReference {
	return n.ObjectReference
}
func (n *FcDataObject) getMmsDataObj() *Data {
	dataStructure := NewDataStructure()
	for _, modelNode := range n.getChildren() {
		child := modelNode.getMmsDataObj()
		if child == nil {
			throw("Unable to convert Child: " + modelNode.getObjectReference().toString() + " to MMS Data Object.")
		}
		dataStructure.seqOf = append(dataStructure.seqOf, child)
	}

	if len(dataStructure.seqOf) == 0 {
		throw("Converting ModelNode: " + n.ObjectReference.toString() + " to MMS Data Object resulted in Sequence of size zero.")
	}

	data := NewData()
	data.structure = dataStructure

	return data
}

func (n *FcDataObject) setValueFromMmsDataObj(data *Data) {
	//	log.Println("FcDataObject")
	if data.structure == nil {
		throw("TYPE_CONFLICT expected type: structure")
	}
	if len(data.structure.seqOf) != len(n.Children) {
		throw(
			"TYPE_CONFLICT  expected type: structure with " + strconv.Itoa(len(n.Children)) + " elements")
	}
	/*index := 0
	for _, child := range n.Children {
		child.setValueFromMmsDataObj(data.structure.seqOf[index])
		index++
	}

	*/

	i := 0
	for _, child := range n.Children {

		if _, ok := child.(*ConstructedDataAttribute); ok {
			//log.Println("ConstructedDataAttribute Data", child.(*ConstructedDataAttribute))
			for _, data1 := range data.structure.seqOf {
				if data1.structure != nil {
					child.setValueFromMmsDataObj(data1)

				}
			}

		}
		if _, ok := child.(*BdaQuality); ok {
			//	log.Println("BdaQuality Data", child.(*BdaQuality))
			for _, data1 := range data.structure.seqOf {
				if data1.bitString != nil {
					child.setValueFromMmsDataObj(data1)

				}

			}

		}
		if _, ok := child.(*BdaBoolean); ok {
			for _, data1 := range data.structure.seqOf {
				if data1.bool != nil {
					child.setValueFromMmsDataObj(data1)
				}
			}

		}
		if _, ok := child.(*BdaInt8U); ok {
			//	log.Println("BdaInt8U Data", child.(*BdaInt8U))
			for _, data1 := range data.structure.seqOf {
				if data1.unsigned != nil {
					child.setValueFromMmsDataObj(data1)

				}

			}

		}
		if _, ok := child.(*BdaInt8); ok {
			for _, data1 := range data.structure.seqOf {
				if data1.integer != nil {
					child.setValueFromMmsDataObj(data1)

				}
			}

		}
		if _, ok := child.(*BdaInt32); ok {
			for _, data1 := range data.structure.seqOf {
				if data1.integer != nil {
					child.setValueFromMmsDataObj(data1)

				}
			}
		}
		if _, ok := child.(*BdaFloat32); ok {
			for _, data1 := range data.structure.seqOf {
				if data1.FloatingPoint != nil {

					child.setValueFromMmsDataObj(data1)

				}

			}

		}
		if _, ok := child.(*BdaFloat64); ok {
			for _, data1 := range data.structure.seqOf {
				if data1.FloatingPoint != nil {

					child.setValueFromMmsDataObj(data1)

				}

			}

		}
		if _, ok := child.(*BdaTimestamp); ok {
			//log.Println("BdaTimestamp Data", child.(*BdaTimestamp))
			for _, data1 := range data.structure.seqOf {
				if data1.utcTime != nil {

					child.setValueFromMmsDataObj(data1)

				}
			}

		}
		if _, ok := child.(*BdaDoubleBitPos); ok {
			//log.Println("BdaDoubleBitPos Data", child.(*BdaDoubleBitPos))
			for _, data1 := range data.structure.seqOf {
				_ = data1
				//	log.Println("FcDataObject", data1)
				/*if data1. != nil {
					child.setValueFromMmsDataObj(data1)
				}

				*/
			}
			return
		}

		i++
	}
}
func remove(slice []*Data, s int) []*Data {
	return append(slice[:s], slice[s+1:]...)
}

func NewFcDataObject(objectReference *ObjectReference, fc string, children []ModelNodeI) *FcDataObject {
	f := &FcDataObject{}
	f.Children = make(map[string]ModelNodeI)
	f.ObjectReference = objectReference
	for _, child := range children {
		f.Children[child.getObjectReference().getName()] = (ModelNodeI)(child)
		child.setParent(f)
	}
	f.Fc = fc

	return f
}
