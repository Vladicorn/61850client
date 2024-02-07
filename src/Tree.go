package src

import (
	"errors"
	"strings"
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

func GetTreeSl(association *ClientAssociation) ([]*Leaf, error) {
	mainTree := make([]*Leaf, 2)
	serverModel, err := association.RetrieveModel()
	if err != nil {
		return nil, err
	}
	dataSets := make(map[string][]*Leaf)

	mainRoot := make([]*Leaf, 0, len(serverModel.getChildren()))
	mainReport := make([]*Leaf, 0, len(serverModel.getChildren()))

	for nameDataset, datasets := range serverModel.DataSets {
		childTree := make([]*Leaf, 0, len(datasets.MembersMap))
		ret := &Leaf{
			Path:   nameDataset,
			Var:    true,
			Type:   IEC61850_VALTYPE_DIRECTORY,
			Childs: childTree,
		}

		for _, members := range datasets.MembersMap {
			for _, mem := range members {
				err = ParseFcDataObjectSl(mem, ret)
				if err != nil {
					return nil, err
				}
			}

		}
		dataSets[nameDataset] = ret.Childs
	}

	for _, lg := range serverModel.getChildren() {

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

					err = ParseFcDataObjectSl(do, retL2)
					if err != nil {
						return nil, err
					}
				case *Urcb:
					tt := do.(*Urcb)
					datasetRef := tt.ObjectReference.toString() + ".DatSet"
					fcModelNode1, err := serverModel.AskForFcModelNode(datasetRef, "RP")
					if err != nil {
						return nil, err
					}
					association.GetDataValues(fcModelNode1)
					ff := fcModelNode1.(*BdaVisibleString)
					itemID := strings.ReplaceAll(string(ff.value), "$", ".")
					retReport := &Leaf{
						Path:   itemID,
						Var:    true,
						Type:   IEC61850_VALTYPE_DIRECTORY,
						Childs: dataSets[itemID],
					}
					mainReport = append(mainReport, retReport)

				case *Brcb:
					tt := do.(*Brcb)
					datasetRef := tt.ObjectReference.toString() + ".DatSet"
					fcModelNode1, err := serverModel.AskForFcModelNode(datasetRef, "BR")
					if err != nil {
						return nil, err
					}
					association.GetDataValues(fcModelNode1)
					ff := fcModelNode1.(*BdaVisibleString)
					itemID := strings.ReplaceAll(string(ff.value), "$", ".")
					retReport := &Leaf{
						Path:   itemID,
						Var:    true,
						Type:   IEC61850_VALTYPE_DIRECTORY,
						Childs: dataSets[itemID],
					}
					mainReport = append(mainReport, retReport)

				default:
					return nil, errors.New("unknown type")
				}
			}
		}
	}

	retMainDataModel := &Leaf{
		Path:   "DataModels",
		Var:    true,
		Type:   IEC61850_VALTYPE_DIRECTORY,
		Childs: mainRoot,
	}

	retMainReports := &Leaf{
		Path:   "Reports",
		Var:    true,
		Type:   IEC61850_VALTYPE_DIRECTORY,
		Childs: mainRoot,
	}
	mainTree[0] = retMainDataModel
	mainTree[1] = retMainReports

	return mainTree, nil
}

func ParseFcDataObjectSl(lgs ModelNodeI, tree *Leaf) error {
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
			err := ParseFcDataObjectSl(lg, ret)
			if err != nil {
				return err
			}
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
			err := ParseFcDataObjectSl(lg, ret)
			if err != nil {
				return err
			}
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
		return errors.New("unknown type")
	}
	return nil
}
