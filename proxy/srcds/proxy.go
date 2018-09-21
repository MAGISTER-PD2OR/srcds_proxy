package srcds

import (
	"net"
	"github.com/bonnetn/srcds_proxy/proxy/config"
	"github.com/bonnetn/srcds_proxy/utils"
	"github.com/golang/glog"
	m "github.com/bonnetn/srcds_proxy/proxy/srcds/model"
	connectionMapper "github.com/bonnetn/srcds_proxy/proxy/srcds/mapper/connection"
)

func Listen(done <-chan utils.DoneEvent, addr string) (*Listener, error) {
	conn, err := makeConnection(done, addr, true)
	if err != nil {
		return nil, err
	}
	return &Listener{
		conn: conn,
	}, err

}

func AssociateWithServerConnection(done <-chan utils.DoneEvent, connChan <-chan m.Connection) <-chan m.Binding {
	result := make(chan m.Binding)
	go func() {
		defer close(result)

		for clientConnection := range connChan {
			if utils.IsDone(done) {
				return
			}

			udpConn, err := dial(done, config.ServerAddr())
			if err != nil {
				return
			}
			glog.V(4).Info("New server connection created.")

			result <- m.Binding{
				ServerConnection: connectionMapper.ToServerConnection(done, udpConn),
				ClientConnection: clientConnection,
			}
		}
	}()
	return result
}

func dial(done <-chan utils.DoneEvent, addr string) (*net.UDPConn, error) {
	return makeConnection(done, addr, false)
}

func makeConnection(done <-chan utils.DoneEvent, addr string, listening bool) (*net.UDPConn, error) {

	// Listen will create a listening UDP ClientConnection.

	// First create the UDP listening ClientConnection.
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	var udpConn *net.UDPConn
	if listening {
		udpConn, err = net.ListenUDP("udp", laddr)
	} else {
		udpConn, err = net.DialUDP("udp", nil, laddr)
	}
	if err != nil {
		return nil, err
	}

	if err = setSRCSConnectionOptions(udpConn); err != nil {
		return nil, err
	}

	// Close on done
	go func() {
		<-done
		udpConn.Close()
	}()

	return udpConn, nil
}

func setSRCSConnectionOptions(connection *net.UDPConn) error {
	// Set the buffers size
	if err := connection.SetWriteBuffer(config.MaxDatagramSize); err != nil {
		return err
	}
	if err := connection.SetReadBuffer(config.MaxDatagramSize); err != nil {
		return err
	}
	return nil
}
