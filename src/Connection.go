package src

import (
	"log"
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

func ConnectToUnBufferReport(association *ClientAssociation, report string, firstGet bool) error {
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
	serverModel, _ := association.RetrieveModel()
	tree := make(map[string]struct{})
	for num, items := range serverModel.Children {
		parceTree(num, items, tree)
	}
}

func parceTree(num string, items ModelNodeI, tree map[string]struct{}) {
	lg, ok := items.(*LogicalDevice)
	if ok {
		for numLg, itemsLg := range lg.Children {
			lg1 := itemsLg.(*LogicalNode)
			for numLg1, itemsLg1 := range lg1.Children {
				//	log.Println(numLg1)
				switch itemsLg1.(type) {
				case *FcDataObject:
					lg2 := itemsLg1.(*FcDataObject)
					for numLg2 := range lg2.Children {
						tree[num+"."+numLg+"."+numLg1+"."+numLg2] = struct{}{}
					}

				case *Urcb:
					lg2 := itemsLg1.(*Urcb)
					for numLg2, _ := range lg2.Children {
						tree[num+"."+numLg+"."+numLg1+"."+numLg2] = struct{}{}
					}

					/*
						aa := lg2.Rcb
						aa1 := aa.GetObjectReference()
						log.Println("TRTRT", aa1)
						//	association.SetDataValues(aa)

					*/
				case *Brcb:
					lg2 := itemsLg1.(*Brcb)
					for numLg2 := range lg2.Children {
						tree[num+"."+numLg+"."+numLg1+"."+numLg2] = struct{}{}
					}

				default:
					log.Println("Unknown")
				}
			}
		}
	}
}
