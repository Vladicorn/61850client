package src

type BasicDataAttribute struct {
	FcModelNode
	basicType string
	sAddr     string
	dchg      bool
	dupd      bool
	chgRcbs   []*Urcb
	dupdRcbs  []*Urcb
}

func NewBasicDataAttribute(objectReference *ObjectReference, fc string, sAddr string, dchg bool, dupd bool) *BasicDataAttribute {
	b := &BasicDataAttribute{}
	b.objectReference = objectReference
	b.Fc = fc
	b.sAddr = sAddr
	b.dchg = dchg
	b.dupd = dupd

	if dchg {
		b.chgRcbs = make([]*Urcb, 0)
	} else {
		b.chgRcbs = nil
	}
	if dupd {
		b.dupdRcbs = make([]*Urcb, 0)
	} else {
		b.dupdRcbs = nil
	}

	return b
}
