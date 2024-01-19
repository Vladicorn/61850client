package main

import (
	"github.com/moxiaolong/61850client/src"
	"log"
	"time"
)

type myClient struct{}

func main() {
	clientSap := src.NewClientSap()
	association := clientSap.Associate("192.168.0.67", 102, src.NewEventListener())

	//getItems(serverModel)

	ticker := time.NewTicker(500 * time.Millisecond)
	tickerCom := time.NewTicker(20 * time.Second)
	readDataSet(association)
	/*	fcModelNode := serverModel.AskForFcModelNode("ied1lDevice1/LLN0.NamPlt.vendor", "DC")
		fcModelNode.(*src.BdaVisibleString).SetValue("fdsfsd")
		association.SetDataValues(fcModelNode)
	*/

	for {
		select {
		case <-tickerCom.C:
		//	ClientTestInfo(client)
		case <-ticker.C:
			/*	err := readValue(association)
				if err != nil {
						log.Println(err)
				}

			*/
		}
	}

}

func getItems(serverModel *src.ServerModel) {
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
					case *src.FcDataObject:
						lg2 := itemsLg1.(*src.FcDataObject)
						for numLg2 := range lg2.Children {
							log.Println(num + "." + numLg + "." + numLg1 + "." + numLg2)
						}
					case *src.Urcb:
						lg2 := itemsLg1.(*src.Urcb)
						for numLg2 := range lg2.Children {
							log.Println(num + "." + numLg + "." + numLg1 + "." + numLg2)
						}
					case *src.Brcb:
						lg2 := itemsLg1.(*src.Brcb)
						for numLg2 := range lg2.Children {
							log.Println(num + "." + numLg + "." + numLg1 + "." + numLg2)
						}

					default:
						log.Println("Unknown")
					}
				}
			}
		}
	}
}
func readDataSet(association *src.ClientAssociation) error {
	serverModel := association.RetrieveModel()

	ff := serverModel.GetDataSet("Demo1ProtCtrl/LLN0.DS2_Protection")

	tt := association.SetDataSetValues(ff)

	log.Println(tt)

	return nil
}
func readValue(association *src.ClientAssociation) error {
	serverModel := association.RetrieveModel()

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
