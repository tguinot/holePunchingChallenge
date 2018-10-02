package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"shared"
)


func SetupChannel(port string) *net.UDPConn {
	myAddressSuf := ":" + port
	udpAddr, err := net.ResolveUDPAddr("udp4", myAddressSuf)
	if err != nil {
		log.Fatal(err)
	}

	UDPChannel, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal(err)
	}

	return UDPChannel
}

func SavePeer(message *shared.Message, address *net.UDPAddr) {
	remoteAddr := fmt.Sprintf("%s:%d", address.IP, address.Port)
	peerMap[message.Header.Peer] = remoteAddr
}

func Acknowledge(message *shared.Message, UDPChannel *net.UDPConn, address *net.UDPAddr) error {
	messageRequest := shared.CraftMessage(2, message.Header.Peer, "HELLO")
	return shared.SendMessageTo(messageRequest, UDPChannel, address)
}

func GivePeerAddress(message *shared.Message, UDPChannel *net.UDPConn, address *net.UDPAddr) error {
	var addressStr string
	if _, ok := peerMap[message.Content]; ok {
		addressStr = peerMap[message.Content]
	}

	messageRequest := shared.CraftMessage(2, message.Header.Peer, addressStr)
	return shared.SendMessageTo(messageRequest, UDPChannel, address)
}


func Respond(message *shared.Message, UDPChannel *net.UDPConn, address *net.UDPAddr) error {
	if message.Header.Purpose == 0 {
		SavePeer(message, address)
		err := Acknowledge(message, UDPChannel, address)
		if err != nil {
			return err
		}
	} else if message.Header.Purpose == 1 {
		err := GivePeerAddress(message, UDPChannel, address)
		if err != nil {
			return err
		}
	}

	return nil
}

func TreatRequests(UDPChannel *net.UDPConn) {
	for {
		message, address, err := shared.Receive(UDPChannel)
		if err != nil {
			log.Print(err)
			return
		}

		err = Respond(message, UDPChannel, address)
		if err != nil {
			log.Print(err)
		}
	}
}

var peerMap map[string]string

func main() {
	peerMap = map[string]string{}
	if len(os.Args) != 2 {
		log.Fatal("usage: ", os.Args[0], "port")
	}

	UDPChannel := SetupChannel(os.Args[1])

	TreatRequests(UDPChannel)
}
