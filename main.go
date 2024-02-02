package main

import (
	"github.com/moxiaolong/61850client/src"
	"log"
)

func main() {

	variableBufReport := []string{"DemoMeasurement/LLN0.brcb1"}
	variableUnBufReport := []string{"DemoMeasurement/LLN0.urcb1", "Bresler43LD1/LLN0.urcbH01", "D001CTRL/LLN0.urcbCTRL_C01", "ied1lDevice1/LLN0.urcb102"}
	var err error
	clientSap := src.NewClientSap()
	event := src.NewClientEventListener()
<<<<<<< Updated upstream
	association := clientSap.Associate("192.168.0.67", 102, event)

	ticker := time.NewTicker(500 * time.Millisecond)

	/*err := subscribeOnDataset(association,
	"DemoMeasurement/LLN0$DS1_Measurement", "DemoMeasurement/LLN0.urcb1",
	true)


	*/
	/*
		err := subscribeOnDataset(association,
			"Bresler1LD1/LLN0$DS_RDRE", "Bresler1LD1/LLN0.urcbC02",
			true)

	*/
	err := subscribeOnDataset(association,
		"D001CTRL/LLN0$ds_A", "D001CTRL/LLN0.urcbCTRL_C01",
		false)

	/*
		err := subscribeOnDataset(association,
			"ied1lDevice1/LLN0$dataset1", "ied1lDevice1/LLN0.urcb102",
			true)

	*/

=======
	association, err := clientSap.Associate("192.168.0.67", 102, event)
>>>>>>> Stashed changes
	if err != nil {
		log.Println(err)
		return
	}

	err = src.ConnectToBufferReport(association,
		variableBufReport[0],
		false)

	err = src.ConnectToUnBufferReport(association,
		variableUnBufReport[0],
		false)

	src.GetTree(association)

	for {
		select {
		case report := <-event.Values:
			log.Println(report)
		}
	}

}

func getItems(association *src.ClientAssociation) {
	serverModel, _ := association.RetrieveModel()
	for num, items := range serverModel.Children {
		//	log.Println(num)
		lg, ok := items.(*src.LogicalDevice)
		if ok {
			for numLg, itemsLg := range lg.Children {
				//	log.Println(numLg)
				lg1 := itemsLg.(*src.LogicalNode)
				for numLg1, itemsLg1 := range lg1.Children {
					//	log.Println(numLg1)
					switch itemsLg1.(type) {
					/*case *src.FcDataObject:
					lg2 := itemsLg1.(*src.FcDataObject)
					for numLg2 := range lg2.Children {
						log.Println(num + "." + numLg + "." + numLg1 + "." + numLg2)
					}

					*/
					case *src.Urcb:
						lg2 := itemsLg1.(*src.Urcb)
						for numLg2, _ := range lg2.Children {
							//	log.Println(tt)
							log.Println(num + "." + numLg + "." + numLg1 + "." + numLg2)

						}

						aa := lg2.Rcb
						aa1 := aa.GetObjectReference()

						log.Println("TRTRT", aa1)
					//	association.SetDataValues(aa)
					/*case *src.Brcb:
					lg2 := itemsLg1.(*src.Brcb)
					for numLg2 := range lg2.Children {
						log.Println(num + "." + numLg + "." + numLg1 + "." + numLg2)
					}

					*/

					default:
						//	log.Println("Unknown")
					}
				}
			}
		}
	}
}

func readDataSet(association *src.ClientAssociation) error {
	//serverModel := association.RetrieveModel()

	//ff := serverModel.GetDataSet("ied1lDevice1/LLN0.dataset1")

	//	ff := serverModel.GetDataSet("Demo1ProtCtrl/Obj2XSWI1.Pos")

	//association.SetDataSetValues(ff)

	/*association.GetDataValues(ff.Members[0])
	tt := ff.Members[0].(*src.FcDataObject)
	log.Println(tt.GetObjectReference())

	association.GetDataValues(ff.Members[1])

	tt1 := ff.Members[1].(*src.BdaFloat32)
	log.Println(tt1.GetValueString())

	*/
	//str := association.SetDataSetValues(ff)
	//data := src.NewDataSet("Demo1ProtCtrl/LLN0.DS2_Protection", ff.Members, true)
	/*
		dataset := src.NewDataSet(ff.DataSetReference, ff.Members, false)

		tt := association.SetDataSetValues(dataset)

	*/

	//log.Println(ff)

	return nil
}

<<<<<<< Updated upstream
/*
func report(association *src.ClientAssociation) {
	log.Println("sets")
	serverModel := association.RetrieveModel()
	//1
	fcModelNode, err := serverModel.AskForFcModelNode("ied1lDevice1/LLN0.urcb101.Resv", "RP")
	if err != nil {
		log.Println(err)
	}
	fcModelNode.(*src.BdaBoolean).SetValue(true)
	association.SetDataValues(fcModelNode)

	//2
	fcModelNode2, err := serverModel.AskForFcModelNode("ied1lDevice1/LLN0.urcb101.DatSet", "RP")
	if err != nil {
		log.Println(err)
	}

	fcModelNode2.(*src.BdaVisibleString).SetValue("ied1lDevice1/LLN0$dataset1")
	association.SetDataValues(fcModelNode2)

	//3
	fcModelNode1, err := serverModel.AskForFcModelNode("ied1lDevice1/LLN0.urcb101.RptEna", "RP")
	if err != nil {
		log.Println(err)
	}
	fcModelNode1.(*src.BdaBoolean).SetValue(true)
	association.SetDataValues(fcModelNode1)

	//4
	fcModelNode4, err := serverModel.AskForFcModelNode("ied1lDevice1/LLN0.urcb101.GI", "RP")
	if err != nil {
		log.Println(err)
	}
	fcModelNode4.(*src.BdaBoolean).SetValue(true)
	association.SetDataValues(fcModelNode4)
	//5
	fcModelNode5, err := serverModel.AskForFcModelNode("ied1lDevice1/LLN0.urcb101.RptEna", "RP")
	if err != nil {
		log.Println(err)
	}
	fcModelNode5.(*src.BdaBoolean).SetValue(true)
	association.SetDataValues(fcModelNode5)

}

*/

func subscribeOnDataset(association *src.ClientAssociation, dataset string, report string, firstGet bool) error {
	// добавить тэг TrgOps - указывает, какие события будут вызывать отчеты. Возможные события:

	serverModel := association.RetrieveModel()
	//1

	fcModelNode, err := serverModel.AskForFcModelNode(fmt.Sprintf("%s.Resv", report), "RP")
	if err != nil {
		return err
	}
	fcModelNode.(*src.BdaBoolean).SetValue(true)
	association.SetDataValues(fcModelNode)

	//2
	fcModelNode2, err := serverModel.AskForFcModelNode(report+".DatSet", "RP")
	if err != nil {
		return err
	}

	fcModelNode2.(*src.BdaVisibleString).SetValue(dataset)
	association.SetDataValues(fcModelNode2)

	//4
	fcModelNode6, err := serverModel.AskForFcModelNode(report+".OptFlds", "RP")
	if err != nil {
		return err
	}
	optops := make([]byte, 2)
	optops[0] = 124
	optops[1] = 0

	fcModelNode6.(*src.BdaOptFlds).SetValue(optops)
	association.SetDataValues(fcModelNode6)

	//3
	fcModelNode1, err := serverModel.AskForFcModelNode(report+".RptEna", "RP")
	if err != nil {
		return err
	}
	fcModelNode1.(*src.BdaBoolean).SetValue(true)
	association.SetDataValues(fcModelNode1)

	/*
		//4
		fcModelNode6, err := serverModel.AskForFcModelNode(report+".TrgOps", "RP")
		if err != nil {
			return err
		}
		trgops := make([]byte, 2)
		trgops[0] = 124
		trgops[1] = 0

		fcModelNode6.(*src.BdaTriggerConditions).SetValue(trgops)
		association.SetDataValues(fcModelNode6)

	*/

	//5
	fcModelNode4, err := serverModel.AskForFcModelNode(report+".GI", "RP")
	if err != nil {
		return err
	}
	fcModelNode4.(*src.BdaBoolean).SetValue(firstGet)
	association.SetDataValues(fcModelNode4)
	//6
	fcModelNode5, err := serverModel.AskForFcModelNode(report+".RptEna", "RP")
	if err != nil {
		return err
	}
	fcModelNode5.(*src.BdaBoolean).SetValue(true)
	association.SetDataValues(fcModelNode5)
	return nil
}

=======
>>>>>>> Stashed changes
func readValue(association *src.ClientAssociation) error {
	serverModel, _ := association.RetrieveModel()

	ttt := serverModel.DataSets
	log.Println(ttt)
	//fcModelNode := serverModel.AskForFcModelNode("ied1lDevice1/MMXU1.TotW.mag.f", "MX")
	//fcModelNode := serverModel.AskForFcModelNode("Demo1ProtCtrl/LLN0.Mod.stVal", "ST")
	fcModelNode1, err := serverModel.AskForFcModelNode("Demo1Measurement/I3pMHAI1.Mod.ctlModel", "CF")
	if err != nil {
		return err
	}
	fcModelNode2, err := serverModel.AskForFcModelNode("Demo1Measurement/LLN0.NamPlt.vendor", "DC")
	if err != nil {
		return err
	}

	association.GetDataValues(fcModelNode1)
	association.GetDataValues(fcModelNode2)
	fcNodeBasic1 := fcModelNode1.(src.BasicDataAttributeI)
	fcNodeBasic2 := fcModelNode2.(src.BasicDataAttributeI)
	println(fcNodeBasic1.GetValueString())
	println(fcNodeBasic2.GetValueString())
	return nil
}
