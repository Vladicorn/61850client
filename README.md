Fork https://github.com/moxiaolong/61850client

Подключение
```go
clientSap := src.NewClientSap()
association := clientSap.Associate(hostName, port, src.NewEventListener())
```
Получение дерева
```go
serverModel := association.RetrieveModel()
```
Получения значения от тэга
```go
fcModelNode := serverModel.AskForFcModelNode("ied1lDevice1/MMXU1.TotW.mag.f", "MX")
association.GetDataValues(fcModelNode)
fcNodeBasic := fcModelNode.(src.BasicDataAttributeI)
println(fcNodeBasic.GetValueString())
```
Запись значения в тэг
```go
fcModelNode := serverModel.AskForFcModelNode("ied1lDevice1/LLN0.NamPlt.vendor", "DC")
fcModelNode.(*src.BdaVisibleString).SetValue("abc")
association.SetDataValues(fcModelNode)
```

	