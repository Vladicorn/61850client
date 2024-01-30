package src

import (
	"log"
	"strconv"
)

type ConstructedDataAttribute struct {
	FcModelNode
}

func (c *ConstructedDataAttribute) getMmsDataObj() *Data {
	structure := NewDataStructure()

	for _, modelNode := range c.getChildren() {
		child := modelNode.getMmsDataObj()
		if child == nil {
			throw(
				"Unable to convert Child: " + modelNode.getObjectReference().toString() + " to MMS Data Object.")
		}
		structure.seqOf = append(structure.seqOf, child)
	}

	if len(structure.seqOf) == 0 {
		throw("Converting ModelNode: " + c.getObjectReference().toString() + " to MMS Data Object resulted in Sequence of size zero.")
	}

	data := NewData()
	data.structure = structure

	return data
}

func (c *ConstructedDataAttribute) setValueFromMmsDataObj(data *Data) {
	//	log.Println("ConstructedDataAttribute", len(data.structure.seqOf))

	if data.structure == nil {
		log.Println("ServiceError.TYPE_CONFLICT expected type: structure")
		return
		throw("ServiceError.TYPE_CONFLICT expected type: structure")
	}
	if len(data.structure.seqOf) != len(c.Children) {
		return
		throw("ServiceError.TYPE_CONFLICT expected type: structure with " + strconv.Itoa(len(c.Children)) + " elements")
	}
	i := 0

	for _, child := range c.Children {
		if _, ok := child.(*ConstructedDataAttribute); ok {
			for _, data1 := range data.structure.seqOf {
				if data1.structure != nil {
					child.setValueFromMmsDataObj(data1)
				}
			}
		}

		if _, ok := child.(*BdaQuality); ok {
			for _, data1 := range data.structure.seqOf {
				if data1.bitString != nil {
					child.setValueFromMmsDataObj(data1)
				}
			}
		}

		if _, ok := child.(*BdaInt8U); ok {
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
					//log.Println("BdaFloat64")
					child.setValueFromMmsDataObj(data1)
				}
			}
		}

		if _, ok := child.(*BdaTimestamp); ok {
			for _, data1 := range data.structure.seqOf {
				if data1.utcTime != nil {
					child.setValueFromMmsDataObj(data1)
				}
			}
		}

		if _, ok := child.(*BdaDoubleBitPos); ok {
			for _, data1 := range data.structure.seqOf {
				log.Println("BdaDoubleBitPos", data1)
			}
		}

		i++
	}

}

func (c *ConstructedDataAttribute) parceStructure(data *Data) {
	if data.bitString != nil {
		i := 0
		for _, child := range c.Children {
			child.setValueFromMmsDataObj(data.structure.seqOf[i])
			i++
		}
	}
	if data.utcTime != nil {
		i := 0
		for _, child := range c.Children {
			child.setValueFromMmsDataObj(data.structure.seqOf[i])
			i++
		}
	}
}

func (c *ConstructedDataAttribute) copy() ModelNodeI {
	subDataAttributesCopy := make([]ModelNodeI, 0)
	for _, subDA := range c.Children {
		subDataAttributesCopy = append(subDataAttributesCopy, subDA.copy())
	}
	return NewConstructedDataAttribute(c.getObjectReference(), c.Fc, subDataAttributesCopy)
}

func NewConstructedDataAttribute(objectReference *ObjectReference, fc string, children []ModelNodeI) *ConstructedDataAttribute {
	c := &ConstructedDataAttribute{}
	c.ObjectReference = objectReference
	c.Fc = fc
	c.Children = make(map[string]ModelNodeI)
	for _, child := range children {
		c.Children[child.getName()] = child
		child.setParent(c)
	}

	return c
}
