package src

import (
	"bytes"
	"container/list"
	"strings"
	"sync"
	"unsafe"
)

type ClientReceiver struct {
	pduBuffer             *bytes.Buffer
	maxMmsPduSize         int
	lock                  *sync.Mutex
	closed                bool
	acseAssociation       *AcseAssociation
	reportListener        *ClientEventListener
	incomingResponses     *list.List
	incomingResponsesLock *sync.Mutex
	expectedResponseId    int
	association           *ClientAssociation
}

func NewClientReceiver(maxMmsPduSize int, association *ClientAssociation) *ClientReceiver {
	return &ClientReceiver{maxMmsPduSize: maxMmsPduSize, closed: false, incomingResponses: list.New(), expectedResponseId: -1, association: association}
}
func (r *ClientReceiver) start() {
	go r.run()
}

func (r *ClientReceiver) run() {
	defer func() {
		err := recover()
		if err != nil {
			r.close(err)
		}
	}()

	for {
		r.pduBuffer.Reset()
		var buffer []byte
		buffer = r.acseAssociation.receive(r.pduBuffer)
		decodedResponsePdu := NewMMSpdu()
		decodedResponsePdu.decode(bytes.NewBuffer(buffer))

		if decodedResponsePdu.unconfirmedPDU != nil {
			if decodedResponsePdu.unconfirmedPDU.Service.informationReport.VariableAccessSpecification.ListOfVariable != nil {
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
			r.incomingResponsesLock.Lock()
			{
				if r.expectedResponseId == -1 {
					// Discarding Reject MMS PDU because no listener for request was found.
					continue
				} else if decodedResponsePdu.rejectPDU.OriginalInvokeID.value != r.expectedResponseId {
					// Discarding Reject MMS PDU because no listener with fitting invokeID was found.
					continue
				} else {
					r.incomingResponses.PushBack(decodedResponsePdu)
				}
			}
			r.incomingResponsesLock.Unlock()
		} else if decodedResponsePdu.confirmedErrorPDU != nil {
			r.incomingResponsesLock.Lock()

			if r.expectedResponseId == -1 {
				// Discarding ConfirmedError MMS PDU because no listener for request was found.
				continue
			} else if decodedResponsePdu.confirmedErrorPDU.invokeID.value != r.expectedResponseId {
				// Discarding ConfirmedError MMS PDU because no listener with fitting invokeID was
				// found.
				continue
			} else {
				r.incomingResponses.PushBack(decodedResponsePdu)
			}
			r.incomingResponsesLock.Unlock()
		} else {
			r.incomingResponsesLock.Lock()

			if r.expectedResponseId == -1 {
				// Discarding ConfirmedResponse MMS PDU because no listener for request was found.
				continue
			} else if decodedResponsePdu.confirmedResponsePDU.invokeID.value != r.expectedResponseId {
				// Discarding ConfirmedResponse MMS PDU because no listener with fitting invokeID
				// was
				// found.
				continue
			} else {
				r.incomingResponses.PushBack(decodedResponsePdu)
			}
			r.incomingResponsesLock.Unlock()
		}

	}

}

func (r *ClientReceiver) close(err any) {
	r.lock.Lock()
	if r.closed == false {
		r.closed = true
		r.acseAssociation.disconnect()

		if r.reportListener != nil {
			go r.reportListener.associationClosed(err)
		}

		mmsPdu := NewMMSpdu()
		mmsPdu.confirmedRequestPDU = NewConfirmedRequestPDU()

		r.incomingResponses.PushBack(mmsPdu)

	}

	r.lock.Unlock()
}

func (r *ClientReceiver) processReport(mmsPdu *MMSpdu) *Report {
	if mmsPdu.unconfirmedPDU == nil {
		throw("getReport: Error decoding server response")
	}

	unconfirmedRes := mmsPdu.unconfirmedPDU

	if unconfirmedRes.Service == nil {
		throw("getReport: Error decoding server response")
	}

	unconfirmedServ := unconfirmedRes.Service

	if unconfirmedServ.informationReport == nil {
		throw("getReport: Error decoding server response")
	}

	listRes :=
		unconfirmedServ.informationReport.listOfAccessResult.seqOf

	index := 0

	if listRes[index].Success.visibleString == nil {
		throw("processReport: report does not contain RptID")
	}
	index++
	rptId := listRes[index].Success.visibleString.toString()

	if listRes[index].Success.bitString == nil {
		throw("processReport: report does not contain OptFlds")
	}

	optFlds := NewBdaOptFlds(NewObjectReference("none"), "")
	index++
	optFlds.value = listRes[(index)].Success.bitString.value

	var sqNum *int = nil
	if optFlds.isSequenceNumber() {
		index++
		sqNum = &listRes[index].Success.Unsigned.value
	}

	var timeOfEntry *BdaEntryTime = nil
	if optFlds.isReportTimestamp() {
		timeOfEntry = NewBdaEntryTime(NewObjectReference("none"), "", "", false, false)
		index++
		timeOfEntry.setValueFromMmsDataObj(listRes[index].Success)
	}

	dataSetRef := ""
	if optFlds.isDataSetName() {
		index++
		dataSetRef = listRes[index].Success.visibleString.toString()
	} else {
		urcbs := r.association.ServerModel.urcbs
		for s := range urcbs {
			urcb := urcbs[s]
			if urcb.getRptId() != nil && urcb.getRptId().getStringValue() == (rptId) || urcb.objectReference.toString() == (rptId) {
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

	dataSet := r.association.ServerModel.getDataSet(dataSetRef)
	if dataSet == nil {
		throw(
			"unable to find data set that matches the given data set reference of the report.")
	}

	var bufOvfl *bool
	if optFlds.isBufferOverflow() {
		index++
		bufOvfl = &listRes[index].Success.bool.value
	}

	var entryId *BdaOctetString = nil
	if optFlds.isEntryId() {
		entryId = NewBdaOctetString(NewObjectReference("none"), "", "", 8, false, false)
		index++
		entryId.setValue(listRes[index].Success.OctetString.value)
	}

	var confRev *int = nil
	if optFlds.isConfigRevision() {

		index++
		confRev = &listRes[index].Success.Unsigned.value
	}

	var subSqNum *int = nil
	moreSegmentsFollow := false
	if optFlds.isSegmentation() {
		index++
		subSqNum = &listRes[index].Success.Unsigned.value
		index++
		moreSegmentsFollow = listRes[index].Success.bool.value
	}

	index++
	inclusionBitString := listRes[index].Success.bitString.getValueAsBooleans()
	numMembersReported := 0

	for _, bit := range inclusionBitString {
		if bit {
			numMembersReported++
		}
	}

	if optFlds.isDataReference() {
		// this is just to move the index to the right place
		// The next part will process the changes to the values
		// without the dataRefs
		index += numMembersReported
	}

	reportedDataSetMembers := make([]*FcModelNode, numMembersReported)
	dataSetIndex := 0
	for _, dataSetMember := range dataSet.getMembers() {
		if inclusionBitString[dataSetIndex] {
			index++
			accessRes := listRes[index]

			c := dataSetMember.copy()
			pointer := unsafe.Pointer(c)
			dataSetMemberCopy := (*FcModelNode)(pointer)
			dataSetMemberCopy.setValueFromMmsDataObj(accessRes.Success)
			reportedDataSetMembers = append(reportedDataSetMembers, dataSetMemberCopy)
		}
		dataSetIndex++
	}

	var reasonCodes []*BdaReasonForInclusion = nil
	if optFlds.isReasonForInclusion() {
		reasonCodes = make([]*BdaReasonForInclusion, len(dataSet.getMembers()))
		for i := 0; i < len(dataSet.getMembers()); i++ {
			if inclusionBitString[i] {

				reasonForInclusion := NewBdaReasonForInclusion(nil)
				reasonCodes = append(reasonCodes, reasonForInclusion)
				index++
				reason := listRes[index].Success.bitString.value
				reasonForInclusion.value = reason
			}
		}
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
		reasonCodes)
}
