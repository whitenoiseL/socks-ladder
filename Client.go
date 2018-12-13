package socks_ladder

import (
	"net"
	"log"
)

type LsLocal struct {
	*SecureSocket
}

func New(password *Password, listenAddr, remoteAddr *net.TCPAddr) *LsLocal{
	return &LsLocal{
		SecureSocket: &SecureSocket{
			Cipher: NewCipher(password),
			ListenAddr: listenAddr,
			RemoteAddr: remoteAddr,
		},
	}
}

func (local *LsLocal) Listen(didListen func(listenAddr net.Addr)) error {
	listener, err := net.ListenTCP("tcp", local.SecureSocket.ListenAddr)
	if err != nil {
		return err
	}

	defer listener.Close()

	if didListen != nil {
		didListen(listener.Addr())
	}

	for {
		userConn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}

		userConn.SetLinger(0)
		go local.handleConn(userConn)
	}
	return nil
}

func (local *LsLocal) handleConn(userConn *net.TCPConn) {
	defer userConn.Close()

	proxyServer, err := local.SecureSocket.DialRemote()
	if err != nil {
		log.Println(err)
		return
	}
	defer proxyServer.Close()

	proxyServer.SetLinger(0)

	go func() {
		err := local.SecureSocket.DecodeCopy(userConn, proxyServer)
		if err != nil {
			userConn.Close()
			proxyServer.Close()
		}
	}()

	local.EncodeCopy(proxyServer, userConn)
}