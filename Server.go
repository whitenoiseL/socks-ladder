package socks_ladder

import (
	"net"
	"log"
	"encoding/binary"
)

type LsServer struct {
	*SecureSocket
}

func NewLsServer(password *Password, listenAddr *net.TCPAddr) *LsServer {
	return &LsServer{
		SecureSocket: &SecureSocket{
			Cipher:     NewCipher(password),
			ListenAddr: listenAddr,
		},
	}
}

func (lsServer *LsServer) Listen(didListen func(listenAddr net.Addr)) error {
	listener, err := net.ListenTCP("tcp", lsServer.ListenAddr)
	if err != nil {
		return err
	}

	defer listener.Close()

	if didListen != nil {
		didListen(listener.Addr())
	}

	for {
		localConn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		localConn.SetLinger(0)
		go lsServer.handleConn(localConn)
	}
	return nil
}

func (lsServer *LsServer) handleConn(localConn *net.TCPConn) {
	defer localConn.Close()
	buf := make([]byte, 256)

	_, err := lsServer.SecureSocket.DecodeRead(localConn, buf)
	if err != nil || buf[0] != 0x05 {
		return
	}

	lsServer.EncodeWrite(localConn, []byte{0x05,0x00})

	n, err := lsServer.SecureSocket.DecodeRead(localConn, buf)
	if err != nil || n < 7 {
		return
	}

	if buf[1] != 0x01 {
		return
	}

	var dIP []byte

	switch buf[3] {
	case 0x01:
		dIP = buf [4:4+net.IPv4len]
	case 0x03:
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:n-2]))
		if err != nil {
			return
		}
		dIP = ipAddr.IP
	case 0x04:
		dIP = buf[4:4+net.IPv6len]
	default:
		return
	}

	dPort := buf[n-2:]
	dstAddr := &net.TCPAddr{
		IP:   dIP,
		Port: int(binary.BigEndian.Uint16(dPort)),
	}
	dstServer, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		return
	} else {
		defer dstServer.Close()
		dstServer.SetLinger(0)

		lsServer.EncodeWrite(localConn, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}

	go func() {
		err := lsServer.DecodeCopy(dstServer, localConn)
		if err != nil {
			localConn.Close()
			dstServer.Close()
		}
	}()

	lsServer.EncodeCopy(localConn, dstServer)
}