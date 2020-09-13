package apollo_client

import (
	"log"
	"net"
)

type ChangeType int

const (
	Add    ChangeType = 0
	Update ChangeType = 1
	Delete ChangeType = 2
)

func initIp() string {
	addr, err := net.ResolveUDPAddr("udp", "8.8.8.8:53")
	if err != nil {
		return ""
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer conn.Close()
	return conn.LocalAddr().String()
}
