package src

import (
	"log"
	"strings"
)

func ConnectToBufferReport(association *ClientAssociation, report string, firstGet bool) error {
	// добавить тэг TrgOps - указывает, какие события будут вызывать отчеты. Возможные события:

	serverModel, err := association.RetrieveModel()
	if err != nil {
		return err
	}

	fcModelNode6, err := serverModel.AskForFcModelNode(report+".TrgOps", "BR")
	if err != nil {
		return err
	}
	TrgOps := make([]byte, 1)
	TrgOps[0] = 124

	fcModelNode6.(*BdaTriggerConditions).SetValue(TrgOps)
	err = association.SetDataValues(fcModelNode6)
	if err != nil {
		return err
	}

	fcModelNode1, err := serverModel.AskForFcModelNode(report+".RptEna", "BR")
	if err != nil {
		return err
	}
	fcModelNode1.(*BdaBoolean).SetValue(true)
	err = association.SetDataValues(fcModelNode1)
	if err != nil {
		return err
	}

	fcModelNode4, err := serverModel.AskForFcModelNode(report+".GI", "BR")
	if err != nil {
		return err
	}
	fcModelNode4.(*BdaBoolean).SetValue(firstGet)
	err = association.SetDataValues(fcModelNode4)
	if err != nil {
		return err
	}

	return nil
}

func ConnectToUnBufferReport(association *ClientAssociation, report string, firstGet bool, nameReport string) error {
	// добавить тэг TrgOps - указывает, какие события будут вызывать отчеты. Возможные события:
	serverModel, _ := association.RetrieveModel()
	//1
	fcModelNode, err := serverModel.AskForFcModelNode(report+".Resv", "RP")
	if err != nil {
		return err
	}
	fcModelNode.(*BdaBoolean).SetValue(true)
	err = association.SetDataValues(fcModelNode)
	if err != nil {
		return err
	}

	//7
	fcModelNode7, err := serverModel.AskForFcModelNode(report+".RptID", "RP")
	if err != nil {
		return err
	}

	fcModelNode7.(*BdaVisibleString).SetValue(nameReport)
	err = association.SetDataValues(fcModelNode7)
	if err != nil {
		return err
	}

	//4
	fcModelNode6, err := serverModel.AskForFcModelNode(report+".OptFlds", "RP")
	if err != nil {
		return err
	}
	optops := make([]byte, 2)
	optops[0] = 124
	optops[1] = 0

	fcModelNode6.(*BdaOptFlds).SetValue(optops)
	err = association.SetDataValues(fcModelNode6)
	if err != nil {
		return err
	}
	//3
	fcModelNode1, err := serverModel.AskForFcModelNode(report+".RptEna", "RP")
	if err != nil {
		return err
	}
	fcModelNode1.(*BdaBoolean).SetValue(true)
	err = association.SetDataValues(fcModelNode1)
	if err != nil {
		return err
	}
	//5
	fcModelNode4, err := serverModel.AskForFcModelNode(report+".GI", "RP")
	if err != nil {
		return err
	}
	fcModelNode4.(*BdaBoolean).SetValue(firstGet)
	err = association.SetDataValues(fcModelNode4)
	if err != nil {
		return err
	}
	//6
	fcModelNode5, err := serverModel.AskForFcModelNode(report+".RptEna", "RP")
	if err != nil {
		return err
	}
	fcModelNode5.(*BdaBoolean).SetValue(true)
	err = association.SetDataValues(fcModelNode5)
	if err != nil {
		return err
	}
	return nil
}

func GetTree(association *ClientAssociation) {
	treeDataModel := make([]map[string]struct{}, 0)
	serverModel, _ := association.RetrieveModel()
	dataSets := make(map[string]map[string]struct{})
	treeReport := make(map[string]map[string]struct{})

	for nameDataset, datasets := range serverModel.DataSets {
		tree1 := make(map[string]struct{})
		for _, members := range datasets.MembersMap {
			for _, mem := range members {

				ParseFcDataObject(mem, tree1)
			}

		}
		dataSets[nameDataset] = tree1
	}

	for _, lg := range serverModel.getChildren() {
		tree := make(map[string]struct{})

		logicalDevice := lg.(*LogicalDevice)
		for _, ln := range logicalDevice.getChildren() {
			logicalnode := ln.(*LogicalNode)
			for _, do := range logicalnode.getChildren() {
				switch do.(type) {
				case *FcDataObject:
					ParseFcDataObject(do, tree)
				case *Urcb:
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
				default:
					log.Println("Unknow")
				}
			}
		}
		treeDataModel = append(treeDataModel, tree)
	}
	for treeRepor := range treeReport {
		log.Println(treeRepor, "    ", treeReport[treeRepor])
	}
}

func ParseFcDataObject(lgs ModelNodeI, tree map[string]struct{}) {
	switch lgs.(type) {
	case *FcDataObject:
		val := lgs.(*FcDataObject)
		logicalNode := val.getChildren()
		for _, lg := range logicalNode {
			ParseFcDataObject(lg, tree)
		}
	case *BdaBoolean:
		val := lgs.(*BdaBoolean)

		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaTimestamp:
		val := lgs.(*BdaTimestamp)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaQuality:
		val := lgs.(*BdaQuality)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *ConstructedDataAttribute:
		val := lgs.(*ConstructedDataAttribute)
		logicalNode := val.getChildren()
		for _, lg := range logicalNode {
			ParseFcDataObject(lg, tree)
		}
		//log.Println(val.getChildren())
		//tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaVisibleString:
		val := lgs.(*BdaVisibleString)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaCheck:
		val := lgs.(*BdaCheck)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaOctetString:
		val := lgs.(*BdaOctetString)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaInt8:
		val := lgs.(*BdaInt8)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaInt8U:
		val := lgs.(*BdaInt8U)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaInt16:
		val := lgs.(*BdaInt16)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaInt16U:
		val := lgs.(*BdaInt16U)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaInt32:
		val := lgs.(*BdaInt32)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaInt32U:
		val := lgs.(*BdaInt32U)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaInt64:
		val := lgs.(*BdaInt64)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaFloat32:
		val := lgs.(*BdaFloat32)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaFloat64:
		val := lgs.(*BdaFloat64)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *FCArray:
		val := lgs.(*FCArray)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaDoubleBitPos:
		val := lgs.(*BdaDoubleBitPos)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaTriggerConditions:
		val := lgs.(*BdaTriggerConditions)
		tree[val.getObjectReference().toString()] = struct{}{}
	case *BdaOptFlds:
		val := lgs.(*BdaOptFlds)
		tree[val.getObjectReference().toString()] = struct{}{}
	default:
		val := lgs.(*BdaQuality)
		log.Println(val)
	}

}
