package src

import (
	"encoding/json"
	"log"
)

type Type61850 int

const (
	IEC61850_VALTYPE_BOOL = Type61850(iota)
	IEC61850_VALTYPE_TIME
	IEC61850_VALTYPE_QUALITY
	IEC61850_VALTYPE_DIRECTORY
)

type Leaf struct {
	Path   string
	Var    bool
	Type   Type61850
	Childs []*Leaf
}

func GetTreeSl(association *ClientAssociation) {
	serverModel, _ := association.RetrieveModel()
	//	treeDataModel := make([]map[string]struct{}, 0)
	//	dataSets := make(map[string]map[string]struct{})
	//	treeReport := make(map[string]map[string]struct{})

	mainRoot := make([]*Leaf, 0, len(serverModel.getChildren()))

	/*for nameDataset, datasets := range serverModel.DataSets {
		tree1 := make(map[string]struct{})
		for _, members := range datasets.MembersMap {
			for _, mem := range members {

				ParseFcDataObjectSl(mem, tree1)
			}

		}
		dataSets[nameDataset] = tree1
	}

	*/

	for _, lg := range serverModel.getChildren() {
		//tree := make(map[string]struct{})

		logicalDevice := lg.(*LogicalDevice)

		childTreeL0 := make([]*Leaf, 0, len(logicalDevice.getChildren()))
		retL0 := &Leaf{
			Path:   logicalDevice.getObjectReference().toString(),
			Var:    true,
			Type:   IEC61850_VALTYPE_DIRECTORY,
			Childs: childTreeL0,
		}
		mainRoot = append(mainRoot, retL0)

		for _, ln := range logicalDevice.getChildren() {
			logicalnode := ln.(*LogicalNode)

			childTreeL1 := make([]*Leaf, 0, len(logicalnode.getChildren()))
			retL1 := &Leaf{
				Path:   logicalnode.getObjectReference().toString(),
				Var:    true,
				Type:   IEC61850_VALTYPE_DIRECTORY,
				Childs: childTreeL1,
			}

			retL0.Childs = append(retL0.Childs, retL1)

			for _, do := range logicalnode.getChildren() {
				switch do.(type) {
				case *FcDataObject:
					fcDataObj := ln.(*LogicalNode)
					childTreeL2 := make([]*Leaf, 0, len(fcDataObj.getChildren()))
					retL2 := &Leaf{
						Path:   fcDataObj.getObjectReference().toString(),
						Var:    true,
						Type:   IEC61850_VALTYPE_DIRECTORY,
						Childs: childTreeL2,
					}

					retL1.Childs = append(retL1.Childs, retL2)

					ParseFcDataObjectSl(do, retL2)
				/*case *Urcb:
					tt := do.(*Urcb)
					datasetRef := tt.ObjectReference.toString() + ".DatSet"
					fcModelNode1, err := serverModel.AskForFcModelNode(datasetRef, "RP")
					if err != nil {
						log.Println(err)
						return
					}
					association.GetDataValues(fcModelNode1)
					ff := fcModelNode1.(*BdaVisibleString)
					itemID := strings.ReplaceAll(string(ff.value), "$", ".")
					treeReport[tt.ObjectReference.toString()] = dataSets[itemID]

				case *Brcb:
					tt := do.(*Brcb)
					datasetRef := tt.ObjectReference.toString() + ".DatSet"
					fcModelNode1, err := serverModel.AskForFcModelNode(datasetRef, "BR")
					if err != nil {
						log.Println(err)
						return
					}
					association.GetDataValues(fcModelNode1)
					ff := fcModelNode1.(*BdaVisibleString)
					itemID := strings.ReplaceAll(string(ff.value), "$", ".")
					treeReport[tt.ObjectReference.toString()] = dataSets[itemID]

				*/
				default:
					//	log.Println("Unknow")
				}
			}
		}

		//treeDataModel = append(treeDataModel, tree)
	}

	byt, _ := json.Marshal(mainRoot)
	log.Println(string(byt))

}

func ParseFcDataObjectSl(lgs ModelNodeI, tree *Leaf) {
	switch lgs.(type) {
	case *FcDataObject:
		val := lgs.(*FcDataObject)
		logicalNode := val.getChildren()

		childTree := make([]*Leaf, 0, len(logicalNode))
		ret := &Leaf{
			Path:   val.getObjectReference().toString(),
			Var:    true,
			Type:   IEC61850_VALTYPE_DIRECTORY,
			Childs: childTree,
		}
		tree.Childs = append(tree.Childs, ret)
		for _, lg := range logicalNode {
			ParseFcDataObjectSl(lg, ret)
		}
	case *BdaBoolean:
		val := lgs.(*BdaBoolean)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaTimestamp:
		val := lgs.(*BdaTimestamp)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_TIME,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaQuality:
		val := lgs.(*BdaQuality)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_QUALITY,
		}
		tree.Childs = append(tree.Childs, ret)
	case *ConstructedDataAttribute:
		val := lgs.(*ConstructedDataAttribute)
		logicalNode := val.getChildren()
		childTree := make([]*Leaf, 0, len(logicalNode))
		ret := &Leaf{
			Path:   val.getObjectReference().toString(),
			Var:    true,
			Type:   IEC61850_VALTYPE_DIRECTORY,
			Childs: childTree,
		}
		tree.Childs = append(tree.Childs, ret)

		for _, lg := range logicalNode {
			ParseFcDataObjectSl(lg, ret)
		}

	case *BdaVisibleString:
		val := lgs.(*BdaVisibleString)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaCheck:
		val := lgs.(*BdaCheck)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaOctetString:
		val := lgs.(*BdaOctetString)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt8:
		val := lgs.(*BdaInt8)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt8U:
		val := lgs.(*BdaInt8U)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt16:
		val := lgs.(*BdaInt16)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt16U:
		val := lgs.(*BdaInt16U)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt32:
		val := lgs.(*BdaInt32)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt32U:
		val := lgs.(*BdaInt32U)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt64:
		val := lgs.(*BdaInt64)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaFloat32:
		val := lgs.(*BdaFloat32)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaFloat64:
		val := lgs.(*BdaFloat64)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *FCArray:
		val := lgs.(*FCArray)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaDoubleBitPos:
		val := lgs.(*BdaDoubleBitPos)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaTriggerConditions:
		val := lgs.(*BdaTriggerConditions)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	case *BdaOptFlds:
		val := lgs.(*BdaOptFlds)
		ret := &Leaf{
			Path: val.getObjectReference().toString(),
			Var:  false,
			Type: IEC61850_VALTYPE_BOOL,
		}
		tree.Childs = append(tree.Childs, ret)
	default:
		val := lgs.(*BdaQuality)
		log.Println(val)
	}

}
