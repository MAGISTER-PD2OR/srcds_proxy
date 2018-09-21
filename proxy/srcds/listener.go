package srcds

import (
	"net"
	"github.com/bonnetn/srcds_proxy/utils"
	"github.com/golang/glog"
	m "github.com/bonnetn/srcds_proxy/proxy/srcds/model"
	connectionMapper "github.com/bonnetn/srcds_proxy/proxy/srcds/mapper/connection"
)

type Listener struct {
	conn *net.UDPConn
}

var clientConnTable m.ConnectionTable

func (l *Listener) Accept(done chan utils.DoneEvent) <-chan m.Connection {
	result := make(chan m.Connection)
	go func() {
		defer close(result)

		buffer := m.GetBufferPool().Get()
		defer m.GetBufferPool().Put(buffer)

		for {
			n, raddr, err := l.conn.ReadFromUDP(buffer)
			if utils.IsDone(done) {
				return
			}
			if err != nil {
				return
			}

			clientConn, loaded := l.getOrCreateClientConn(done, raddr)
			if !loaded {
				glog.V(1).Info("Packet received with no connection assigned, creating new connection.")
				result <- clientConn.Connection
				glog.V(1).Info("Connection created.")
			}
			msg := m.GetBufferPool().Get()
			copy(msg, buffer[:n])
			glog.V(3).Info("Received datagram of length ", n, " from a client.")
			clientConn.MsgChan <- msg[:n]
			glog.V(3).Info("Forwarded datagram of length ", n, " in the input channel.")
		}

	}()
	return result
}

func (l *Listener) getOrCreateClientConn(done <-chan utils.DoneEvent, raddr *net.UDPAddr) (*m.ConnectionWrapper, bool) {
	// Create a new connection.
	killNewClientConn := make(chan utils.DoneEvent)
	newClientConn := connectionMapper.ToClientConnectionWrapper(channelOr(done, killNewClientConn), l.conn, *raddr)

	clientConn, loaded := clientConnTable.GetOrReplace(m.UDPAddrToAddressPort(*raddr), newClientConn)
	if loaded {
		close(killNewClientConn) // If this connection is not used, kill the workers related to that connection.
	}
	return clientConn, loaded
}

func channelOr(a, b <-chan utils.DoneEvent) <-chan utils.DoneEvent {
	orChan := make(chan utils.DoneEvent)
	go func() {
		defer close(orChan)
		for {
			select {
			case <-a:
				return
			case <-b:
				return
			}
		}
	}()
	return orChan
}
