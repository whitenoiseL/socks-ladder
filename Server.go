package socks_ladder

import "net"

type LsServer struct {
	*SecureSocket
}

func New(password *Password, listenAddr *net.TCPAddr) *LsServer {
	return &LsServer{
		SecureSocket: &SecureSocket{
			Cipher:     NewCipher(password),
			ListenAddr: listenAddr,
		},
	}
}

