package ring

import "loadbalance/types"

type ring struct {
	next *ring
	Conn types.Conn
}

func new(conn types.Conn) *ring {
	if nil == conn {
		return nil
	}

	sr := &ring{
		Conn: conn,
	}
	sr.next = sr

	return sr
}

func (r *ring) Next() *ring {
	return r.next
}

func (r *ring) Delete() {
	r.next = r.next.next
}

func (r *ring) Link(s *ring) {
	if s != nil {
		s.next = r.next
		r.next = s
	}
}
