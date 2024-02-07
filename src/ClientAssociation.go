package src

import "C"
import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type ClientAssociation struct {
	ServerModel           *ServerModel
	responseTimeout       int
	negotiatedMaxPduSize  int
	reportListener        *ClientEventListener
	AcseAssociation       *AcseAssociation
	clientReceiver        *ClientReceiver
	servicesSupported     []byte
	lock                  *sync.Mutex
	Closed                bool
	incomingResponses     chan *MMSpdu
	incomingResponsesLock *sync.Mutex
	invokeId              int
	reverseOStream        *ReverseByteArrayOutputStream
}

func NewClientAssociation(address string, port int, acseSap *ClientAcseSap, proposedMaxPduSize int,
	proposedMaxServOutstandingCalling int, proposedMaxServOutstandingCalled int, proposedDataStructureNestingLevel int,
	servicesSupportedCalling []byte, responseTimeout int, messageFragmentTimeout int, reportListener *ClientEventListener) (*ClientAssociation, error) {

	var err error
	c := &ClientAssociation{}
	c.lock = &sync.Mutex{}
	c.incomingResponses = make(chan *MMSpdu)
	c.incomingResponsesLock = &sync.Mutex{}
	c.Closed = false
	c.responseTimeout = responseTimeout
	acseSap.tSap.MessageFragmentTimeout = messageFragmentTimeout
	acseSap.tSap.MessageTimeout = responseTimeout
	c.negotiatedMaxPduSize = proposedMaxPduSize
	c.reportListener = reportListener
	c.reverseOStream, err = NewReverseByteArrayOutputStream(500)
	if err != nil {
		return nil, err
	}

	initiateRequestMMSpdu :=
		constructInitRequestPdu(
			proposedMaxPduSize,
			proposedMaxServOutstandingCalling,
			proposedMaxServOutstandingCalled,
			proposedDataStructureNestingLevel,
			servicesSupportedCalling)

	reverseOStream, err := NewReverseByteArrayOutputStream(500)
	if err != nil {
		return nil, err
	}
	_, err = initiateRequestMMSpdu.encode(reverseOStream)
	if err != nil {
		return nil, err
	}

	c.AcseAssociation, err =
		acseSap.associate(
			address,
			port,
			reverseOStream.getByteBuffer())
	if err != nil {
		return nil, err
	}

	initResponse := c.AcseAssociation.getAssociateResponseAPdu()

	initiateResponseMmsPdu := NewMMSpdu()

	initiateResponseMmsPdu.decode(initResponse)

	err = c.handleInitiateResponse(
		initiateResponseMmsPdu,
		proposedMaxPduSize,
		proposedMaxServOutstandingCalling,
		proposedMaxServOutstandingCalled,
		proposedDataStructureNestingLevel)
	if err != nil {
		return nil, err
	}

	c.AcseAssociation.MessageTimeout = 0
	c.clientReceiver = NewClientReceiver(c.negotiatedMaxPduSize, c)
	err = c.clientReceiver.start()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *ClientAssociation) handleInitiateResponse(responsePdu *MMSpdu, proposedMaxPduSize int, proposedMaxServOutstandingCalling int, proposedMaxServOutstandingCalled int, proposedDataStructureNestingLevel int) error {
	if responsePdu.initiateErrorPDU != nil {
		return errors.New("got response error of class: ") //responsePdu.initiateErrorPDU.errorClass) TODO
	}

	if responsePdu.initiateResponsePDU == nil {
		c.AcseAssociation.disconnect()
		return errors.New("error decoding InitiateResponse Pdu")
	}

	initiateResponsePDU := responsePdu.initiateResponsePDU

	if initiateResponsePDU.localDetailCalled != nil {
		c.negotiatedMaxPduSize = initiateResponsePDU.localDetailCalled.intValue()
	}

	negotiatedMaxServOutstandingCalling :=
		initiateResponsePDU.negotiatedMaxServOutstandingCalling.intValue()

	negotiatedMaxServOutstandingCalled :=
		initiateResponsePDU.negotiatedMaxServOutstandingCalled.intValue()

	var negotiatedDataStructureNestingLevel int
	if initiateResponsePDU.negotiatedDataStructureNestingLevel != nil {
		negotiatedDataStructureNestingLevel =
			initiateResponsePDU.negotiatedDataStructureNestingLevel.intValue()
	} else {
		negotiatedDataStructureNestingLevel = proposedDataStructureNestingLevel
	}

	if c.negotiatedMaxPduSize < 64 || c.negotiatedMaxPduSize > proposedMaxPduSize || negotiatedMaxServOutstandingCalling > proposedMaxServOutstandingCalling || negotiatedMaxServOutstandingCalling < 0 || negotiatedMaxServOutstandingCalled > proposedMaxServOutstandingCalled || negotiatedMaxServOutstandingCalled < 0 || negotiatedDataStructureNestingLevel > proposedDataStructureNestingLevel || negotiatedDataStructureNestingLevel < 0 {

		c.AcseAssociation.disconnect()
		return errors.New("error negotiating parameters")
	}

	version :=
		initiateResponsePDU.initResponseDetail.negotiatedVersionNumber.intValue()
	if version != 1 {
		return errors.New("unsupported version number was negotiated")
	}

	c.servicesSupported = initiateResponsePDU.initResponseDetail.servicesSupportedCalled.value
	if (c.servicesSupported[0] & 0x40) != 0x40 {
		return errors.New("obligatory services are not supported by the server")
	}
	return nil
}

func (c *ClientAssociation) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.Closed == false {
		c.Closed = true
		go c.AcseAssociation.disconnect()
		//go c.reportListener.associationClosed()

		//mmsPdu := NewMMSpdu()
		//mmsPdu.confirmedRequestPDU = NewConfirmedRequestPDU()
		//c.incomingResponses <- mmsPdu
	}

}

func (c *ClientAssociation) RetrieveModel() (*ServerModel, error) {
	ldNames, err := c.retrieveLogicalDevices()
	if err != nil {
		return nil, err
	}
	lnNames := make([][]string, 0)

	for i := 0; i < len(ldNames); i++ {
		lng, err := c.retrieveLogicalNodeNames(ldNames[i])
		if err != nil {
			return nil, err
		}
		lnNames = append(lnNames, lng)
	}

	lds := make([]*LogicalDevice, 0)
	for i := 0; i < len(ldNames); i++ {
		lns := make([]ModelNodeI, 0)
		for j := 0; j < len(lnNames[i]); j++ {
			lns = append(lns, c.retrieveDataDefinitions(
				NewObjectReference(ldNames[i]+"/"+lnNames[i][j])))

		}
		lds = append(lds, NewLogicalDevice(NewObjectReference(ldNames[i]), lns))
	}

	c.ServerModel = NewServerModel(lds, nil)
	err = c.updateDataSets()
	if err != nil {
		return nil, err
	}
	return c.ServerModel, nil

}

func (c *ClientAssociation) retrieveLogicalDevices() ([]string, error) {
	serviceRequest := c.constructGetServerDirectoryRequest()
	confirmedServiceResponse := c.encodeWriteReadDecode(serviceRequest)
	return c.decodeGetServerDirectoryResponse(confirmedServiceResponse)
}

func (c *ClientAssociation) updateDataSets() error {
	if c.ServerModel == nil {
		return errors.New("before calling this function you have to get the ServerModel using the retrieveModel() function")
	}
	lds := c.ServerModel.Children
	for _, ld := range lds {
		serviceRequest :=
			c.constructGetDirectoryRequest(ld.getObjectReference().getName(), "", false)
		confirmedServiceResponse := c.encodeWriteReadDecode(serviceRequest)

		err := c.decodeAndRetrieveDsNamesAndDefinitions(confirmedServiceResponse, ld.(*LogicalDevice))
		if err != nil {
			return err
		}

	}
	return nil
}

func (c *ClientAssociation) retrieveDataDefinitions(lnRef *ObjectReference) *LogicalNode {
	serviceRequest := c.constructGetDataDefinitionRequest(lnRef)
	confirmedServiceResponse := c.encodeWriteReadDecode(serviceRequest)

	return decodeGetDataDefinitionResponse(confirmedServiceResponse, lnRef)
}

func decodeGetDataDefinitionResponse(confirmedServiceResponse *ConfirmedServiceResponse, lnRef *ObjectReference) *LogicalNode {
	return parseGetDataDefinitionResponse(confirmedServiceResponse, lnRef)
}

func (c *ClientAssociation) encodeWriteReadDecode(serviceRequest *ConfirmedServiceRequest) *ConfirmedServiceResponse {
	if c == nil {
		return nil
	}
	currentInvokeId := c.getInvokeId()

	confirmedRequestPdu := NewConfirmedRequestPDU()
	confirmedRequestPdu.invokeID = NewUnsigned32(currentInvokeId)
	confirmedRequestPdu.service = serviceRequest

	requestPdu := NewMMSpdu()
	requestPdu.confirmedRequestPDU = confirmedRequestPdu

	c.reverseOStream.reset()

	func() {
		requestPdu.encode(c.reverseOStream)
	}()

	c.clientReceiver.expectedResponseId = currentInvokeId

	c.AcseAssociation.sendByteBuffer(c.reverseOStream.getByteBuffer())

	var decodedResponsePdu *MMSpdu = nil

	func() {

		if c.responseTimeout == 0 {
			if len(c.incomingResponses) > 0 {
				decodedResponsePdu = <-c.incomingResponses

			}
		} else {
			timeOut := time.After(time.Duration(c.responseTimeout) * time.Millisecond)
			select {
			case decodedResponsePdu = <-c.incomingResponses:
				break
			case <-timeOut:
				panic("time out")
			}
		}
	}()

	if decodedResponsePdu == nil {
		decodedResponsePdu = c.clientReceiver.removeExpectedResponse()
		if decodedResponsePdu == nil {
			throw("Service error TIMEOUT_ERROR")
		}
	}

	if decodedResponsePdu.confirmedRequestPDU != nil {
		c.incomingResponses <- decodedResponsePdu
		throw("connection was Closed", c.clientReceiver.lastIOException)
	}

	testForInitiateErrorResponse(decodedResponsePdu)
	testForErrorResponse(decodedResponsePdu)
	testForRejectResponse(decodedResponsePdu)

	confirmedResponsePdu := decodedResponsePdu.confirmedResponsePDU
	if confirmedResponsePdu == nil {
		throw("Response PDU is not a confirmed response pdu")
	}

	return confirmedResponsePdu.service

}

func testForRejectResponse(mmsResponsePdu *MMSpdu) {
	if mmsResponsePdu.rejectPDU == nil {
		return
	}

	rejectReason := mmsResponsePdu.rejectPDU.rejectReason
	if rejectReason != nil {
		if rejectReason.pduError != nil {
			if rejectReason.pduError.value == 1 {
				throw(
					" PARAMETER_VALUE_INCONSISTENTMMS reject: type: \"pdu-error\", reject code: \"invalid-pdu\"")
			}
		}
	}
	throw(" UNKNOWN MMS confirmed error.")
}

func testForErrorResponse(mmsResponsePdu *MMSpdu) {
	if mmsResponsePdu.confirmedErrorPDU == nil {
		return
	}

	errClass := mmsResponsePdu.confirmedErrorPDU.serviceError.errorClass

	if errClass != nil {
		if errClass.access != nil {
			if errClass.access.value == 3 {
				throw(
					"ACCESS_VIOLATION MMS confirmed error: class: \"access\", error code: \"object-access-denied\"")
			} else if errClass.access.value == 2 {
				throw(
					" INSTANCE_NOT_AVAILABLEMMS confirmed error: class: \"access\", error code: \"object-non-existent\"")
			}

		} else if errClass.file != nil {
			if errClass.file.value == 7 {
				throw(
					"FILE_NONE_EXISTENT  MMS confirmed error: class: \"file\", error code: \"file-non-existent\"")
			}
		}
	}

	if mmsResponsePdu.confirmedErrorPDU.serviceError.additionalDescription != nil {
		throw(
			"UNKNOWN MMS confirmed error. Description: ",
			mmsResponsePdu.confirmedErrorPDU.serviceError.additionalDescription.toString())
	}
	throw("UNKNOWN  MMS confirmed error.")
}

func testForInitiateErrorResponse(mmsResponsePdu *MMSpdu) {
	if mmsResponsePdu.initiateResponsePDU != nil {

		errClass := mmsResponsePdu.initiateErrorPDU.errorClass
		if errClass != nil {
			if errClass.vmdState != nil {
				throw("FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"vmd_state\" with val: ", strconv.Itoa(errClass.vmdState.value))
			}
			if errClass.applicationReference != nil {
				throw("FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"application_reference\" with val: ", strconv.Itoa(errClass.applicationReference.value))
			}
			if errClass.definition != nil {
				throw("FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"definition\" with val: ", strconv.Itoa(errClass.definition.value))
			}
			if errClass.resource != nil {
				throw(
					" FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"resource\" with val: ", strconv.Itoa(errClass.resource.value))
			}
			if errClass.service != nil {
				throw("FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"service\" with val: ", strconv.Itoa(errClass.service.value))
			}
			if errClass.servicePreempt != nil {
				throw(

					"FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT  error class \"service_preempt\" with val: " + strconv.Itoa(errClass.servicePreempt.value))
			}
			if errClass.timeResolution != nil {
				throw(

					"FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"time_resolution\" with val: " + strconv.Itoa(errClass.timeResolution.value))
			}
			if errClass.access != nil {
				throw(
					"FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"access\" with val: " + strconv.Itoa(errClass.access.value))
			}
			if errClass.initiate != nil {
				throw(
					"FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"initiate\" with val: " + strconv.Itoa(errClass.initiate.value))
			}
			if errClass.conclude != nil {
				throw(
					"FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"conclude\" with val: " + strconv.Itoa(errClass.conclude.value))
			}
			if errClass.cancel != nil {
				throw("FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"cancel\" with val: ", strconv.Itoa(errClass.cancel.value))
			}
			if errClass.file != nil {
				throw(

					"FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"file\" with val: " + strconv.Itoa(errClass.file.value))
			}
			if errClass.others != nil {
				throw(
					"FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT error class \"others\" with val: " + strconv.Itoa(errClass.others.value))
			}
		}
		throw(
			"FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT unknown error class")
	}
}

func (c *ClientAssociation) retrieveLogicalNodeNames(ld string) ([]string, error) {
	lns := make([]string, 0)
	var err error
	continueAfterRef := ""
	once := false
	for !once || continueAfterRef != "" {
		once = true
		serviceRequest := c.constructGetDirectoryRequest(ld, continueAfterRef, true)
		confirmedServiceResponse := c.encodeWriteReadDecode(serviceRequest)
		continueAfterRef, lns, err = c.decodeGetDirectoryResponse(confirmedServiceResponse, lns)
		if err != nil {
			return nil, err
		}
	}

	return lns, nil
}

func (c *ClientAssociation) constructGetServerDirectoryRequest() *ConfirmedServiceRequest {
	objectClass := NewObjectClass()
	objectClass.basicObjectClass = NewBerInteger(nil, 9)

	objectScope := NewObjectScope()
	objectScope.vmdSpecific = NewBerNull()

	getNameListRequest := NewGetNameListRequest()
	getNameListRequest.objectClass = objectClass
	getNameListRequest.objectScope = objectScope

	confirmedServiceRequest := NewConfirmedServiceRequest()
	confirmedServiceRequest.getNameList = getNameListRequest

	return confirmedServiceRequest
}

func (c *ClientAssociation) decodeGetServerDirectoryResponse(confirmedServiceResponse *ConfirmedServiceResponse) ([]string, error) {
	objectRefs := make([]string, 0) // ObjectReference[identifiers.size()];
	if confirmedServiceResponse == nil {
		return objectRefs, errors.New("confirmedServiceResponse is nil")
	}
	if confirmedServiceResponse.getNameList == nil {
		return objectRefs, errors.New("FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINTError decoding Get Server Directory Response Pdu")
	}

	identifiers := confirmedServiceResponse.getNameList.listOfIdentifier.getIdentifier()

	for _, identifier := range identifiers {
		objectRefs = append(objectRefs, identifier.toString())
	}

	return objectRefs, nil
}

func (c *ClientAssociation) constructGetDirectoryRequest(ldRef string, continueAfter string, logicalDevice bool) *ConfirmedServiceRequest {

	objectClass := NewObjectClass()

	if logicalDevice {
		objectClass.basicObjectClass = NewBerInteger(nil, 0)
	} else { // for data sets
		objectClass.basicObjectClass = NewBerInteger(nil, 2)
	}

	ldRefByte := *(*[]byte)(unsafe.Pointer(&ldRef))
	objectScopeChoiceType := NewObjectScope()
	objectScopeChoiceType.domainSpecific = NewIdentifier(ldRefByte)

	getNameListRequest := NewGetNameListRequest()
	getNameListRequest.objectClass = objectClass
	getNameListRequest.objectScope = objectScopeChoiceType
	if continueAfter != "" {
		continueAfterByte := *(*[]byte)(unsafe.Pointer(&continueAfter))
		getNameListRequest.continueAfter = NewIdentifier(continueAfterByte)
	}

	confirmedServiceRequest := NewConfirmedServiceRequest()
	confirmedServiceRequest.getNameList = getNameListRequest
	return confirmedServiceRequest
}

func (c *ClientAssociation) decodeAndRetrieveDsNamesAndDefinitions(confirmedServiceResponse *ConfirmedServiceResponse, ld *LogicalDevice) error {
	if confirmedServiceResponse.getNameList == nil {
		return errors.New("serviceError decodeGetDataSetResponse: Error decoding server response")
	}

	getNameListResponse := confirmedServiceResponse.getNameList

	identifiers := getNameListResponse.listOfIdentifier.getIdentifier()

	if len(identifiers) == 0 {
		return nil
	}

	for _, identifier := range identifiers {
		// TODO delete DataSets that no longer exist
		c.getDataSetDirectory(identifier, ld)
	}

	if getNameListResponse.moreFollows != nil && getNameListResponse.moreFollows.value == true {
		return errors.New("FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT")
	}
	return nil
}

func (c *ClientAssociation) constructGetDataDefinitionRequest(lnRef *ObjectReference) *ConfirmedServiceRequest {
	domainSpec := NewDomainSpecific()
	domainSpec.domainID = NewIdentifier([]byte(lnRef.get(0)))
	domainSpec.itemID = NewIdentifier([]byte(lnRef.get(1)))

	objectName := NewObjectName()
	objectName.domainSpecific = domainSpec

	getVariableAccessAttributesRequest := NewGetVariableAccessAttributesRequest()
	getVariableAccessAttributesRequest.name = objectName

	confirmedServiceRequest := NewConfirmedServiceRequest()
	confirmedServiceRequest.getVariableAccessAttributes = getVariableAccessAttributesRequest

	return confirmedServiceRequest
}

func (c *ClientAssociation) getInvokeId() int {
	c.invokeId = (c.invokeId + 1) % 2147483647
	return c.invokeId
}

func (c *ClientAssociation) decodeGetDirectoryResponse(confirmedServiceResponse *ConfirmedServiceResponse, lns []string) (string, []string, error) {
	if confirmedServiceResponse.getNameList == nil {
		return "", lns, errors.New("FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT decodeGetLDDirectoryResponse: Error decoding server response")
	}

	getNameListResponse := confirmedServiceResponse.getNameList

	identifiers := getNameListResponse.listOfIdentifier.getIdentifier()

	if len(identifiers) == 0 {
		return "", lns, errors.New("INSTANCE_NOT_AVAILABLE decodeGetLDDirectoryResponse: Instance not available")
	}

	var identifier *Identifier = nil
	for _, identifier = range identifiers {

		idString := identifier.toString()

		if strings.Index(idString, "$") == -1 {
			lns = append(lns, idString)
		}
	}

	if getNameListResponse.moreFollows != nil && getNameListResponse.moreFollows.value == false {
		return "", lns, nil
	} else {
		return identifier.toString(), lns, nil
	}
}

func (c *ClientAssociation) getDataSetDirectory(dsId *Identifier, ld *LogicalDevice) {
	serviceRequest := c.constructGetDataSetDirectoryRequest(dsId, ld)
	confirmedServiceResponse := c.encodeWriteReadDecode(serviceRequest)
	c.decodeGetDataSetDirectoryResponse(confirmedServiceResponse, dsId, ld)
}

func (c *ClientAssociation) constructGetDataSetDirectoryRequest(dsId *Identifier, ld *LogicalDevice) *ConfirmedServiceRequest {
	domainSpecificObjectName := NewDomainSpecific()
	name := ld.getName()
	nameByte := *(*[]byte)(unsafe.Pointer(&name))
	domainSpecificObjectName.domainID = NewIdentifier(nameByte)
	domainSpecificObjectName.itemID = dsId

	dataSetObj := NewGetNamedVariableListAttributesRequest()
	dataSetObj.domainSpecific = domainSpecificObjectName

	confirmedServiceRequest := NewConfirmedServiceRequest()
	confirmedServiceRequest.getNamedVariableListAttributes = dataSetObj

	return confirmedServiceRequest
}

func (c *ClientAssociation) decodeGetDataSetDirectoryResponse(confirmedServiceResponse *ConfirmedServiceResponse, dsId *Identifier, ld *LogicalDevice) {
	if confirmedServiceResponse.getNamedVariableListAttributes == nil {
		throw(
			"FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT decodeGetDataSetDirectoryResponse: Error decoding server response")
	}

	getNamedVariableListAttResponse :=
		confirmedServiceResponse.getNamedVariableListAttributes
	deletable := getNamedVariableListAttResponse.mmsDeletable.value
	variables :=
		getNamedVariableListAttResponse.listOfVariable.seqOf

	if len(variables) == 0 {
		throw(
			"INSTANCE_NOT_AVAILABLE decodeGetDataSetDirectoryResponse: Instance not available")
	}

	dsMems := make([]FcModelNodeI, 0)

	for _, variableDef := range variables {
		var member FcModelNodeI = nil
		// TODO remove this try catch statement once all possible FCs are
		// supported
		// it is only there so that Functional Constraints such as GS will
		// be ignored and DataSet cotaining elements with these FCs are
		// ignored and not created.
		func() {

			member = c.ServerModel.getNodeFromVariableDef(variableDef).(FcModelNodeI)
		}()

		if member == nil {
			throw(

				"INSTANCE_NOT_AVAILABLE decodeGetDataSetDirectoryResponse: data set memeber does not exist, you might have to call retrieveModel first")
		}
		dsMems = append(dsMems, member)
	}

	dsObjRef := ld.getName() + "/" + strings.ReplaceAll(dsId.toString(), "$", ".")

	dataSet := NewDataSetWithRef(dsObjRef, dsMems, deletable)

	index := strings.Index(dsId.toString(), "$")
	if ld.getChild(dsId.toString()[0:index], "") == nil {
		throw(
			"INSTANCE_NOT_AVAILABLE decodeGetDataSetDirectoryResponse: LN for returned DataSet is not available")
	}

	existingDs := c.ServerModel.GetDataSet(dsObjRef)
	if existingDs == nil {
		c.ServerModel.addDataSet(dataSet)
	} else if !existingDs.deletable {
		return
	} else {
		c.ServerModel.removeDataSet(dsObjRef)
		c.ServerModel.addDataSet(dataSet)
	}
}

func (c *ClientAssociation) GetDataValues(modelNode FcModelNodeI) {
	serviceRequest := c.constructGetDataValuesRequest(modelNode)
	confirmedServiceResponse := c.encodeWriteReadDecode(serviceRequest)
	c.decodeGetDataValuesResponse(confirmedServiceResponse, modelNode)
}

func (c *ClientAssociation) constructGetDataValuesRequest(modelNode FcModelNodeI) *ConfirmedServiceRequest {
	varAccessSpec := c.constructVariableAccessSpecification(modelNode)

	readRequest := NewReadRequest()
	readRequest.variableAccessSpecification = varAccessSpec

	confirmedServiceRequest := NewConfirmedServiceRequest()
	confirmedServiceRequest.read = readRequest

	return confirmedServiceRequest
}

func (c *ClientAssociation) decodeGetDataValuesResponse(confirmedServiceResponse *ConfirmedServiceResponse, modelNode ModelNodeI) {
	if confirmedServiceResponse.read == nil {
		throw(
			"FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT Error decoding GetDataValuesReponsePdu")
	}

	listOfAccessResults := confirmedServiceResponse.read.listOfAccessResult.seqOf

	if len(listOfAccessResults) != 1 {
		throw(
			"PARAMETER_VALUE_INAPPROPRIATE Multiple results received.")
	}

	accRes := listOfAccessResults[0]

	if accRes.failure != nil {
		throw("mmsDataAccessErrorToServiceError", strconv.Itoa(accRes.failure.intValue()))
	}
	modelNode.setValueFromMmsDataObj(accRes.success)
}

func (c *ClientAssociation) constructVariableAccessSpecification(modelNode FcModelNodeI) *VariableAccessSpecification {
	listOfVariable := NewVariableDefs()

	listOfVariable.seqOf = append(listOfVariable.seqOf, modelNode.getMmsVariableDef())

	variableAccessSpecification := NewVariableAccessSpecification()
	variableAccessSpecification.listOfVariable = listOfVariable

	return variableAccessSpecification
}

func (c *ClientAssociation) SetDataValues(node FcModelNodeI) error {
	serviceRequest := c.constructSetDataValuesRequest(node)
	confirmedServiceResponse := c.encodeWriteReadDecode(serviceRequest)
	err := c.decodeSetDataValuesResponse(confirmedServiceResponse)
	if err != nil {
		return err
	}
	return nil
}

// return error string
func (c *ClientAssociation) SetDataSetValues(dataSet *DataSet) []string {
	serviceRequest := c.constructSetDataSetValues(dataSet)
	confirmedServiceResponse := c.encodeWriteReadDecode(serviceRequest)
	return c.decodeSetDataSetValuesResponse(confirmedServiceResponse)
}

func (c *ClientAssociation) constructSetDataValuesRequest(modelNode FcModelNodeI) *ConfirmedServiceRequest {
	variableAccessSpecification := c.constructVariableAccessSpecification(modelNode)

	listOfData := NewListOfData()

	/*
		data := *modelNode.getMmsDataObj()
		data.bool.value = true
	*/

	listOfData.seqOf = append(listOfData.seqOf, modelNode.getMmsDataObj())

	writeRequest := NewWriteRequest()
	writeRequest.listOfData = listOfData

	writeRequest.variableAccessSpecification = variableAccessSpecification

	confirmedServiceRequest := NewConfirmedServiceRequest()
	confirmedServiceRequest.write = writeRequest

	return confirmedServiceRequest
}

func (c *ClientAssociation) decodeSetDataValuesResponse(confirmedServiceResponse *ConfirmedServiceResponse) error {
	writeResponse := confirmedServiceResponse.write

	if writeResponse == nil {
		return errors.New("FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT SetDataValuesResponse: improper response")
	}

	subChoice := writeResponse.seqOf[0]

	if subChoice.failure != nil {
		return errors.New("mmsDataAccessErrorToServiceError" + strconv.Itoa(subChoice.failure.intValue()))
	}

	return nil
}

func (c *ClientAssociation) constructSetDataSetValues(dataSet *DataSet) *ConfirmedServiceRequest {

	varAccessSpec := NewVariableAccessSpecification()
	varAccessSpec.variableListName = dataSet.getMmsObjectName()

	listOfData := NewListOfData()

	for _, member := range dataSet.Members {
		listOfData.seqOf = append(listOfData.seqOf, member.getMmsDataObj())
	}

	writeRequest := NewWriteRequest()
	writeRequest.variableAccessSpecification = varAccessSpec
	writeRequest.listOfData = listOfData

	confirmedServiceRequest := NewConfirmedServiceRequest()
	confirmedServiceRequest.write = writeRequest

	return confirmedServiceRequest
}

func (c *ClientAssociation) decodeSetDataSetValuesResponse(confirmedServiceResponse *ConfirmedServiceResponse) []string {
	if confirmedServiceResponse.write == nil {
		throw("FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT Error decoding SetDataSetValuesReponsePdu")
	}

	writeResponse := confirmedServiceResponse.write
	serviceErrors := make([]string, 0)
	for _, accessResult := range writeResponse.seqOf {
		if accessResult.success != nil {
			serviceErrors = append(serviceErrors, "")
		} else {
			serviceErrors = append(serviceErrors, mmsDataAccessErrorToServiceError(accessResult.failure))
		}
	}

	return serviceErrors
}

func mmsDataAccessErrorToServiceError(dataAccessError *DataAccessError) string {
	switch dataAccessError.value {
	case 1:
		return "FAILED_DUE_TO_SERVER_CONSTRAINT, MMS DataAccessError: hardware-fault"
	case 2:
		return "INSTANCE_LOCKED_BY_OTHER_CLIENT, MMS DataAccessError: temporarily-unavailable"
	case 3:
		return "ACCESS_VIOLATION MMS, DataAccessError: object-access-denied"
	case 5:
		return "PARAMETER_VALUE_INCONSISTENT, MMS DataAccessError: invalid-address"
	case 7:
		return "TYPE_CONFLICT,MMS DataAccessError: type-inconsistent"
	case 10:
		return "INSTANCE_NOT_AVAILABLE, MMS DataAccessError: object-non-existent"
	case 11:
		return "PARAMETER_VALUE_INCONSISTENT, MMS DataAccessError: object-value-invalid"
	default:
		return "FAILED_DUE_TO_COMMUNICATIONS_CONSTRAINT, MMS DataAccessError: " + strconv.Itoa(dataAccessError.value)
	}
}
