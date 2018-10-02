package shared

import (
	"net"
	"encoding/json"
	"log"
)

type Header struct {
	Purpose int
	Peer    string
}
type Message struct {
	Header  Header
	Content string
}

func SendMessageTo(message Message, UDPChannel *net.UDPConn, serverAddress *net.UDPAddr) error {
	jsonRequest, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = UDPChannel.WriteToUDP(jsonRequest, serverAddress)
	return err
}

func CraftMessage(Purpose int, Peer string, Content string) Message {
	return Message{
		Header{Purpose,Peer},
		Content,
	}
}

func Receive(UDPChannel *net.UDPConn) (*Message, *net.UDPAddr, error){
	var recvBuffer [4096]byte
	n, address, err := UDPChannel.ReadFromUDP(recvBuffer[0:])
	if err != nil {
		return nil, nil, err
	}

	var message Message
	err = json.Unmarshal(recvBuffer[:n], &message)
	if err != nil {
		log.Print(err)
		return nil, nil, err
	}

	return &message, address, nil
}