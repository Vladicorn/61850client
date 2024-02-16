package src

import (
	"errors"
	"strings"
)

var IEC61850_VALTYPE_BOOL = 1
var IEC61850_VALTYPE_TIME = 3
var IEC61850_VALTYPE_QUALITY = 3
var IEC61850_VALTYPE_NUMBER = 2
var IEC61850_VALTYPE_STRING = 3
var IEC61850_VALTYPE_DIRECTORY = 0

type Leaf struct {
	Path         string
	Var          bool
	Type         int
	TypeProtocol Iec61850ValueType
	FC           string
	Childs       []*Leaf
}

// return slice with tree level and map without tree
func GetTreeSl(association *ClientAssociation) ([]*Leaf, map[string]*Leaf, error) {
	mainTree := make([]*Leaf, 2)
	serverModel, err := association.RetrieveModel()
	if err != nil {
		return nil, nil, err
	}
	dataSets := make(map[string][]*Leaf)

	mainRoot := make([]*Leaf, 0, len(serverModel.getChildren()))
	mainReport := make([]*Leaf, 0, len(serverModel.getChildren()))
	existRoots := make(map[string]*Leaf)

	for nameDataset, datasets := range serverModel.DataSets {
		childTree := make([]*Leaf, 0, len(datasets.MembersMap))
		ret := &Leaf{
			Path:   nameDataset,
			Var:    false,
			Type:   IEC61850_VALTYPE_DIRECTORY,
			Childs: childTree,
		}

		for _, members := range datasets.MembersMap {
			for _, mem := range members {
				err = ParseFcDataObjectSl(mem, ret, existRoots, "")
				if err != nil {
					return nil, nil, err
				}
			}

		}
		dataSets[nameDataset] = ret.Childs
	}

	existRoots = make(map[string]*Leaf)
	for _, lg := range serverModel.getChildren() {

		logicalDevice := lg.(*LogicalDevice)

		childTreeL0 := make([]*Leaf, 0, len(logicalDevice.getChildren()))
		retL0 := &Leaf{
			Path:   logicalDevice.getObjectReference().toString(),
			Var:    false,
			Type:   IEC61850_VALTYPE_DIRECTORY,
			Childs: childTreeL0,
		}
		mainRoot = append(mainRoot, retL0)

		for _, ln := range logicalDevice.getChildren() {
			logicalnode := ln.(*LogicalNode)

			childTreeL1 := make([]*Leaf, 0, len(logicalnode.getChildren()))
			retL1 := &Leaf{
				Path:   logicalnode.getObjectReference().toString(),
				Var:    false,
				Type:   IEC61850_VALTYPE_DIRECTORY,
				Childs: childTreeL1,
			}

			retL0.Childs = append(retL0.Childs, retL1)
			for _, do := range logicalnode.getChildren() {
				switch do.(type) {
				case *FcDataObject:
					fcDataObj := do.(*FcDataObject)
					if existRoot, ok := existRoots[fcDataObj.getObjectReference().toString()]; !ok {
						childTreeL2 := make([]*Leaf, 0, len(fcDataObj.getChildren()))
						retL2 := &Leaf{
							Path:         fcDataObj.getObjectReference().toString(),
							Var:          false,
							Type:         IEC61850_VALTYPE_DIRECTORY,
							Childs:       childTreeL2,
							FC:           fcDataObj.Fc,
							TypeProtocol: IEC61850_VALUE_TYPE_FcDataObject,
						}
						existRoots[fcDataObj.getObjectReference().toString()] = retL2
						retL1.Childs = append(retL1.Childs, retL2)
						err = ParseFcDataObjectSl(do, retL2, existRoots, "")
						if err != nil {
							return nil, nil, err
						}
					} else {
						err = ParseFcDataObjectSl(fcDataObj, existRoot, existRoots, "")
						if err != nil {
							return nil, nil, err
						}
					}
				case *Urcb:
					tt := do.(*Urcb)
					datasetRef := tt.ObjectReference.toString() + ".DatSet"
					fcModelNode1, err := serverModel.AskForFcModelNode(datasetRef, "RP")
					if err != nil {
						return nil, nil, err
					}
					association.GetDataValues(fcModelNode1)
					ff := fcModelNode1.(*BdaVisibleString)

					itemID := strings.ReplaceAll(string(ff.value), "$", ".")

					existRoots1 := make(map[string]*Leaf)
					if datasets, ok := serverModel.DataSets[itemID]; ok {
						logdevice := strings.Split(itemID, "/")
						name := strings.Split(datasets.DataSetReference, ".")
						childTree := make([]*Leaf, 0, len(datasets.MembersMap))
						ret := &Leaf{
							Path:   name[0],
							Var:    false,
							Type:   IEC61850_VALTYPE_DIRECTORY,
							Childs: childTree,
						}
						existRoots[tt.getName()+"@"+logdevice[0]] = ret

						for _, members := range datasets.MembersMap {
							for _, mem := range members {
								err = ParseFcDataObjectSl(mem, ret, existRoots1, tt.getName()+"@")
								if err != nil {
									return nil, nil, err
								}
							}

						}

						retReport := &Leaf{
							Path:   tt.getName(),
							Var:    false,
							Type:   IEC61850_VALTYPE_DIRECTORY,
							Childs: ret.Childs,
							FC:     tt.Fc,
						}
						mainReport = append(mainReport, retReport)
					}

				case *Brcb:
					tt := do.(*Brcb)
					datasetRef := tt.ObjectReference.toString() + ".DatSet"
					fcModelNode1, err := serverModel.AskForFcModelNode(datasetRef, "BR")
					if err != nil {
						return nil, nil, err
					}
					association.GetDataValues(fcModelNode1)
					ff := fcModelNode1.(*BdaVisibleString)

					itemID := strings.ReplaceAll(string(ff.value), "$", ".")

					existRoots1 := make(map[string]*Leaf)
					if datasets, ok := serverModel.DataSets[itemID]; ok {
						name := strings.Split(datasets.DataSetReference, ".")
						childTree := make([]*Leaf, 0, len(datasets.MembersMap))
						ret := &Leaf{
							Path:   name[0],
							Var:    false,
							Type:   IEC61850_VALTYPE_DIRECTORY,
							Childs: childTree,
						}
						existRoots[tt.getName()] = ret

						for _, members := range datasets.MembersMap {
							for _, mem := range members {
								err = ParseFcDataObjectSl(mem, ret, existRoots1, tt.getName()+"@")
								if err != nil {
									return nil, nil, err
								}
							}

						}

						retReport := &Leaf{
							Path:   tt.getName(),
							Var:    false,
							Type:   IEC61850_VALTYPE_DIRECTORY,
							Childs: ret.Childs,
							FC:     tt.Fc,
						}
						mainReport = append(mainReport, retReport)
					}

				default:
					return nil, nil, errors.New("unknown type")
				}
			}
		}
	}

	retMainDataModel := &Leaf{
		Path:   "DataModels",
		Var:    false,
		Type:   IEC61850_VALTYPE_DIRECTORY,
		Childs: mainRoot,
	}

	retMainReports := &Leaf{
		Path:   "Reports",
		Var:    false,
		Type:   IEC61850_VALTYPE_DIRECTORY,
		Childs: mainReport,
	}
	mainTree[0] = retMainDataModel
	mainTree[1] = retMainReports

	return mainTree, existRoots, nil
}

func ParseFcDataObjectSl(lgs ModelNodeI, tree *Leaf, existRoots map[string]*Leaf, prefix string) error {
	switch lgs.(type) {
	case *FcDataObject:
		val := lgs.(*FcDataObject)
		logicalNode := val.getChildren()

		var ret *Leaf
		if existRoot, ok := existRoots[val.getObjectReference().toString()]; !ok {
			childTree := make([]*Leaf, 0, len(logicalNode))
			ret = &Leaf{
				Path:         prefix + val.getObjectReference().toString(),
				Var:          false,
				Type:         IEC61850_VALTYPE_DIRECTORY,
				Childs:       childTree,
				FC:           val.Fc,
				TypeProtocol: IEC61850_VALUE_TYPE_FcDataObject,
			}
			tree.Childs = append(tree.Childs, ret)
			existRoots[val.getObjectReference().toString()] = ret
		} else {
			ret = existRoot
		}

		for _, lg := range logicalNode {
			err := ParseFcDataObjectSl(lg, ret, existRoots, prefix)
			if err != nil {
				return err
			}
		}
	case *BdaBoolean:
		val := lgs.(*BdaBoolean)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_BOOL,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaBoolean,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaTimestamp:
		val := lgs.(*BdaTimestamp)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_TIME,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaTimestamp,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaQuality:
		val := lgs.(*BdaQuality)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_QUALITY,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaQuality,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *ConstructedDataAttribute:
		val := lgs.(*ConstructedDataAttribute)
		logicalNode := val.getChildren()
		//log.Println(val.getMmsVariableDef().tag)
		var ret *Leaf
		if existRoot, ok := existRoots[val.getObjectReference().toString()]; !ok {
			childTree := make([]*Leaf, 0, len(logicalNode))
			ret = &Leaf{
				Path:         prefix + val.getObjectReference().toString(),
				Var:          false,
				Type:         IEC61850_VALTYPE_DIRECTORY,
				Childs:       childTree,
				FC:           val.Fc,
				TypeProtocol: IEC61850_VALUE_TYPE_ConstructedDataAttribute,
			}
			tree.Childs = append(tree.Childs, ret)
			existRoots[val.getObjectReference().toString()] = ret
		} else {
			ret = existRoot
		}

		for _, lg := range logicalNode {
			err := ParseFcDataObjectSl(lg, ret, existRoots, prefix)
			if err != nil {
				return err
			}
		}

	case *BdaVisibleString:
		val := lgs.(*BdaVisibleString)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_STRING,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaVisibleString,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaCheck:
		val := lgs.(*BdaCheck)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_STRING,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaCheck,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaOctetString:
		val := lgs.(*BdaOctetString)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_STRING,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaOctetString,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt8:
		val := lgs.(*BdaInt8)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_NUMBER,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaInt8,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt8U:
		val := lgs.(*BdaInt8U)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_NUMBER,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaInt8U,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt16:
		val := lgs.(*BdaInt16)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_NUMBER,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaInt16,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt16U:
		val := lgs.(*BdaInt16U)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_NUMBER,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaInt16U,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt32:
		val := lgs.(*BdaInt32)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_NUMBER,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaInt32,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt32U:
		val := lgs.(*BdaInt32U)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_NUMBER,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaInt32U,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaInt64:
		val := lgs.(*BdaInt64)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_NUMBER,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaInt64,
		}

		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaFloat32:
		val := lgs.(*BdaFloat32)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_NUMBER,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaFloat32,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaFloat64:
		val := lgs.(*BdaFloat64)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_NUMBER,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaFloat64,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *FCArray:
		val := lgs.(*FCArray)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_NUMBER,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_FCArray,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaDoubleBitPos:
		val := lgs.(*BdaDoubleBitPos)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_STRING,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaDoubleBitPos,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaTriggerConditions:
		val := lgs.(*BdaTriggerConditions)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_STRING,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaTriggerConditions,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	case *BdaOptFlds:
		val := lgs.(*BdaOptFlds)
		ret := &Leaf{
			Path:         prefix + val.getObjectReference().toString(),
			Var:          true,
			Type:         IEC61850_VALTYPE_STRING,
			FC:           val.Fc,
			TypeProtocol: IEC61850_VALUE_TYPE_BdaOptFlds,
		}
		existRoots[val.getObjectReference().toString()] = ret
		tree.Childs = append(tree.Childs, ret)
	default:
		return errors.New("unknown type")
	}
	return nil
}
