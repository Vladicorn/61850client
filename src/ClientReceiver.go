package src

import (
	"bytes"
	"log"
	"strings"
	"sync"
)

type ClientReceiver struct {
	pduBuffer          *bytes.Buffer
	maxMmsPduSize      int
	lock               *sync.Mutex
	closed             bool
	reportListener     *ClientEventListener
	expectedResponseId int
	association        *ClientAssociation
	lastIOException    string
}

func NewClientReceiver(maxMmsPduSize int, association *ClientAssociation) *ClientReceiver {
	return &ClientReceiver{maxMmsPduSize: maxMmsPduSize, closed: false, reportListener: association.reportListener, expectedResponseId: -1, association: association, pduBuffer: bytes.NewBuffer(make([]byte, maxMmsPduSize+400)),
		lock: &sync.Mutex{}}
}
func (r *ClientReceiver) start() {
	go r.run()
}

func (r *ClientReceiver) run() {
	/*defer func() {
		err := recover()
		if err != nil {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer func() {
					recover()
					wg.Done()
				}()
				r.close(err)
			}()
			wg.Wait()
			fmt.Printf("线程退出 %v \n", err)
		}
	}()

	*/

	for {
		r.pduBuffer.Reset()
		var buffer []byte
		buffer = r.association.AcseAssociation.receive(r.pduBuffer)
		decodedResponsePdu := NewMMSpdu()
		decodedResponsePdu.decode(bytes.NewBuffer(buffer))

		if decodedResponsePdu.unconfirmedPDU != nil {

			if decodedResponsePdu.unconfirmedPDU.service.informationReport.variableAccessSpecification.listOfVariable != nil {
				// Discarding LastApplError Report
			} else {

				if r.reportListener != nil {
					report := r.processReport(decodedResponsePdu)
					go func() {
						r.reportListener.newReport(report)
					}()

				} else {
					// discarding report because no ReportListener was registered.
				}
			}
		} else if decodedResponsePdu.rejectPDU != nil {

			func() {
				defer r.association.incomingResponsesLock.Unlock()
				r.association.incomingResponsesLock.Lock()
				if r.expectedResponseId == -1 {
					// Discarding Reject MMS PDU because no listener for request was found.
					return
				} else if decodedResponsePdu.rejectPDU.originalInvokeID.value != r.expectedResponseId {
					// Discarding Reject MMS PDU because no listener with fitting invokeID was found.
					return
				} else {
					r.association.incomingResponses <- decodedResponsePdu
				}
			}()

		} else if decodedResponsePdu.confirmedErrorPDU != nil {
			func() {
				defer r.association.incomingResponsesLock.Unlock()
				r.association.incomingResponsesLock.Lock()
				if r.expectedResponseId == -1 {
					// Discarding ConfirmedError MMS PDU because no listener for request was found.
					return
				} else if decodedResponsePdu.confirmedErrorPDU.invokeID.value != r.expectedResponseId {
					// Discarding ConfirmedError MMS PDU because no listener with fitting invokeID was
					// found.
					return
				} else {
					r.association.incomingResponses <- decodedResponsePdu
				}
			}()

		} else {
			func() {
				defer r.association.incomingResponsesLock.Unlock()
				r.association.incomingResponsesLock.Lock()
				if r.expectedResponseId == -1 {
					// Discarding ConfirmedResponse MMS PDU because no listener for request was found.
					return
				} else if decodedResponsePdu.confirmedResponsePDU.invokeID.value != r.expectedResponseId {
					// Discarding ConfirmedResponse MMS PDU because no listener with fitting invokeID
					// was
					// found.
					return
				} else {
					r.association.incomingResponses <- decodedResponsePdu
				}
			}()

		}

	}

}

func (r *ClientReceiver) close(err any) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.closed == false {
		r.closed = true
		r.association.AcseAssociation.disconnect()

		if r.reportListener != nil {
			go r.reportListener.associationClosed(err)
		}

		mmsPdu := NewMMSpdu()
		mmsPdu.confirmedRequestPDU = NewConfirmedRequestPDU()

		r.association.incomingResponses <- mmsPdu

	}

}

func (r *ClientReceiver) processReport(mmsPdu *MMSpdu) *Report {
	if mmsPdu.unconfirmedPDU == nil {
		throw("getReport: Error decoding server response")
	}

	unconfirmedRes := mmsPdu.unconfirmedPDU

	if unconfirmedRes.service == nil {
		throw("getReport: Error decoding server response")
	}

	unconfirmedServ := unconfirmedRes.service

	if unconfirmedServ.informationReport == nil {
		throw("getReport: Error decoding server response")
	}

	listRes := unconfirmedServ.informationReport.listOfAccessResult.seqOf
	index := 0

	if listRes[index].success.visibleString == nil {
		throw("processReport: report does not contain RptID")
	}

	//RPTID - visible-string идентификатор отчета идентифицирует RCB, вызвавший создание отчета. Оно соответствует полю RptID RCB.
	rptId := listRes[index].success.visibleString.toString()

	//OptFlds - bit-string равно полю OptFlds соответствующего RCB.
	index++
	if listRes[index].success.bitString == nil {
		throw("processReport: report does not contain OptFlds")
	}
	optFlds := NewBdaOptFlds(NewObjectReference("none"), "")

	optFlds.value = listRes[(index)].success.bitString.value

	// SqNum порядковый номер отчета. На момент отправки оно равно полю SqNum соответствующего RCB.
	//index++
	var sqNum *int = nil
	if optFlds.isSequenceNumber() {
		log.Println("isSequenceNumber")
		index++
		sqNum = &listRes[index].success.unsigned.value
	}

	//TimeOfEntry (необязательно, включается, если OptFlds.report-timestamp имеет значение true) — указывает время создания идентификатора записи.
	var timeOfEntry *BdaEntryTime = nil

	if optFlds.isReportTimestamp() {
		index++
		timeOfEntry = NewBdaEntryTime(NewObjectReference("none"), "", "", false, false)
		timeOfEntry.setValueFromMmsDataObj(listRes[index].success)
	}

	// DatSet (необязательно, включается, если OptFlds.dataset-name имеет значение true) — ссылка на набор данных, данные которого отправляются в этом отчете.
	dataSetRef := ""

	if optFlds.isDataSetName() {
		index++
		dataSetRef = listRes[index].success.visibleString.toString()
	} else {
		urcbs := r.association.ServerModel.urcbs
		for s := range urcbs {
			urcb := urcbs[s]
			if urcb.getRptId() != nil && urcb.getRptId().getStringValue() == (rptId) || urcb.ObjectReference.toString() == (rptId) {
				dataSetRef = urcb.getDatSet().getStringValue()
				break
			}
		}
	}

	if dataSetRef == "" {
		throw(
			"unable to find RCB that matches the given RptID in the report.")
	}

	dataSetRef = strings.ReplaceAll(dataSetRef, "$", ".")

	dataSet := r.association.ServerModel.GetDataSet(dataSetRef)
	if dataSet == nil {
		throw(
			"unable to find data set that matches the given data set reference of the report.")
	}

	var bufOvfl *bool
	if optFlds.isBufferOverflow() {
		log.Println("bufOvfl")
		index++
		//TODO
		if listRes[index].success.bool != nil {
			bufOvfl = &listRes[index].success.bool.value
		}
	}

	var entryId *BdaOctetString = nil
	if optFlds.isEntryId() {
		log.Println("entityId")
		entryId = NewBdaOctetString(NewObjectReference("none"), "", "", 8, false, false)
		index++
		entryId.setValue(listRes[index].success.octetString.value)
	}

	var confRev *int = nil
	if optFlds.isConfigRevision() {
		log.Println("confRev")
		index++
		confRev = &listRes[index].success.unsigned.value
	}

	var subSqNum *int = nil
	moreSegmentsFollow := false
	if optFlds.isSegmentation() {
		log.Println("subSqNum")
		index++
		subSqNum = &listRes[index].success.unsigned.value
		index++
		moreSegmentsFollow = listRes[index].success.bool.value
	}

	index++

	inclusionBitString := listRes[index].success.bitString.bitCheck()
	numMembersReported := 0
	//log.Println("inclusionBitString", listRes[index].success.bitString)
	for _, bit := range inclusionBitString {
		if bit {
			numMembersReported++
		}
	}

	dataReference := make([]string, len(inclusionBitString))
	if optFlds.isDataReference() {
		// this is just to move the index to the right place
		// The next part will process the changes to the values
		// without the dataRefs
		index += numMembersReported
		/*for ind := 0; ind < len(inclusionBitString); ind++ {
			index++
			dataReference[ind] = strings.ReplaceAll(listRes[index].success.visibleString.toString(), "$", ".")
		}

		*/
	}

	reportedDataSetMembers := make([]FcModelNodeI, 0)
	reportedDataSetMembersMap := make(map[string]FcModelNodeI)
	//reportedDataSetMembers := make([]*FcModelNode, numMembersReported)
	dataSetIndex := 7
	index++

	for _, dataSetMember := range dataSet.getMembers() {
		if inclusionBitString[dataSetIndex] {
			accessRes := listRes[index]
			//log.Println("access ", accessRes.success.structure.seqOf)

			//TPDO
			//dataSetMemberCopy := dataSetMember.copy()
			log.Println(accessRes.success)
			dataSetMember.setValueFromMmsDataObj(accessRes.success)
			reportedDataSetMembers = append(reportedDataSetMembers, dataSetMember.(FcModelNodeI))
			reportedDataSetMembersMap[strings.ReplaceAll(listRes[index-numMembersReported].success.visibleString.toString(), "$", ".")] = dataSetMember.(FcModelNodeI)

			index++
		}

		dataSetIndex--
	}

	var reasonCodes []*BdaReasonForInclusion = nil
	if optFlds.isReasonForInclusion() {
		/*
			//reasonCodes = make([]*BdaReasonForInclusion, len(DataSets.getMembers()))
			reasonCodes = make([]*BdaReasonForInclusion, 0)
			for i := 0; i < len(dataSet.getMembers()); i++ {
				if inclusionBitString[i] {
					reasonForInclusion := NewBdaReasonForInclusion(nil)
					reasonCodes = append(reasonCodes, reasonForInclusion)
					index++
					reason := listRes[index].success.bitString.value
					reasonForInclusion.value = reason

				}
			}

		*/

	}

	return NewReport(
		rptId,
		sqNum,
		subSqNum,
		moreSegmentsFollow,
		dataSetRef,
		bufOvfl,
		confRev,
		timeOfEntry,
		entryId,
		inclusionBitString,
		reportedDataSetMembers,
		reasonCodes,
		dataReference,
		reportedDataSetMembersMap,
	)
}

func (r *ClientReceiver) removeExpectedResponse() *MMSpdu {
	//defer r.association.incomingResponsesLock.Unlock()
	//r.association.incomingResponsesLock.Lock()
	r.expectedResponseId = -1
	spdu := <-r.association.incomingResponses
	return spdu
}
