package src

import (
	"bytes"
)

type ClientAcseSap struct {
	tSap *ClientTSap
}

func (s *ClientAcseSap) associate(address string, port int, apdu *bytes.Buffer) (*AcseAssociation, error) {

	a := NewAcseAssociation(nil, []byte{0, 0, 0, 1})

	err := a.startAssociation(
		apdu,
		address,
		port,
		[]byte{0, 1},
		[]byte{0, 1},
		[]byte{0, 0, 0, 1},
		s.tSap,
		[]int{1, 1, 999, 1, 1},
		[]int{1, 1, 999, 1},
		12,
		12)
	if err != nil {
		return nil, err
	}

	defer func() {
		r := recover()
		if r != nil {
			a.disconnect()
			panic(r)
		}
	}()
	return a, nil
}

func newClientAcseSap() *ClientAcseSap {
	c := &ClientAcseSap{}
	c.tSap = NewClientTSap()
	return c
}
