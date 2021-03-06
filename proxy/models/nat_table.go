package models

import (
	"net"
	"sync"
)

// NatTable stores translation between Host objects and UDPConn objects.
type NatTable struct {
	sync.Map
}

// LoadConnection returns the UDPConn associated with the host.
func (tbl *NatTable) LoadConnection(host Host) (conn *net.UDPConn, ok bool) {
	value, ok := tbl.Load(host)
	if !ok {
		return nil, false
	}

	conn, ok = value.(*net.UDPConn)
	return
}

// LoadOrStoreConnection returns the UDPConn associated with the host or store a new connection.
func (tbl *NatTable) LoadOrStoreConnection(host Host, conn *net.UDPConn) (*net.UDPConn, bool) {
	value, loaded := tbl.LoadOrStore(host, conn)
	newConn := value.(*net.UDPConn)
	return newConn, loaded
}
