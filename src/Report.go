package src

type Report struct {
	rptId                     string
	sqNum                     *int
	subSqNum                  *int
	moreSegmentsFollow        bool
	dataSetRef                string
	bufOvfl                   *bool
	confRev                   *int
	timeOfEntry               *BdaEntryTime
	entryId                   *BdaOctetString
	inclusionBitString        []bool
	reportedDataSetMembers    []FcModelNodeI
	reasonCodes               []*BdaReasonForInclusion
	dataReferences            []string
	reportedDataSetMembersMap map[string]FcModelNodeI
}

func NewReport(rptId string, sqNum *int, subSqNum *int, moreSegmentsFollow bool, dataSetRef string,
	bufOvfl *bool, confRev *int, timeOfEntry *BdaEntryTime, entryId *BdaOctetString, inclusionBitString []bool,
	reportedDataSetMembers []FcModelNodeI, reasonCodes []*BdaReasonForInclusion, dataReferences []string, reportedDataSetMembersMap map[string]FcModelNodeI) *Report {
	return &Report{
		rptId:                     rptId,
		sqNum:                     sqNum,
		subSqNum:                  subSqNum,
		moreSegmentsFollow:        moreSegmentsFollow,
		dataSetRef:                dataSetRef,
		bufOvfl:                   bufOvfl,
		confRev:                   confRev,
		timeOfEntry:               timeOfEntry,
		entryId:                   entryId,
		inclusionBitString:        inclusionBitString,
		reportedDataSetMembers:    reportedDataSetMembers,
		reasonCodes:               reasonCodes,
		dataReferences:            dataReferences,
		reportedDataSetMembersMap: reportedDataSetMembersMap,
	}
}
