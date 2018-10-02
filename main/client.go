package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"shared"
)



func SetupNetwork(serverAddr string, port string) (*net.UDPAddr, *net.UDPConn, error) {
	serverAddress, err := net.ResolveUDPAddr("udp4", serverAddr)
	if err != nil {
		log.Print("Error while resolving", serverAddr)
		return nil, nil, err
	}

	addr, err := net.ResolveUDPAddr("udp4", port)
	if err != nil {
		log.Print("Error while resolving", port, err)
		return nil, nil, err
	}
	UDPChannel, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Print("Error while attempting to listen", addr, err)
		return nil, nil, err
	}
	return serverAddress, UDPChannel, err
}

func UpdateServer(UDPChannel *net.UDPConn, nick string, serverAddress *net.UDPAddr) error {
	initMessage := shared.CraftMessage(0, nick, "metadata")
	err := shared.SendMessageTo(initMessage, UDPChannel, serverAddress)
	if err != nil {
		return err
	}

	response, _, err := shared.Receive(UDPChannel)
	if err != nil {
		return err
	}

	fmt.Println("Server says:", response.Content)
	return err
}

func RequestTargetAddress(UDPChannel *net.UDPConn, UDPChannelectMessage shared.Message, serverAddress *net.UDPAddr) (shared.Message, error) {
	recvBuffer := make([]byte, 4096)
	var serverResponse shared.Message

	err := shared.SendMessageTo(UDPChannelectMessage, UDPChannel, serverAddress)
	n, _, err := UDPChannel.ReadFromUDP(recvBuffer)
	if err != nil {
		log.Print("Failed to retrieve target address")
		return serverResponse, err
	}
	err = json.Unmarshal(recvBuffer[:n], &serverResponse)
	return serverResponse, err
}

func LookupTarget(UDPChannel *net.UDPConn, nick string, target string, serverAddress *net.UDPAddr) string {
	UDPChannelectMessage := shared.CraftMessage(1, nick, target)

	serverResponse, err := RequestTargetAddress(UDPChannel, UDPChannelectMessage, serverAddress)
	if err != nil {
		log.Fatal("Error while requesting target address")
	}

	return serverResponse.Content
}

func KeepSending(UDPChannel *net.UDPConn, nick string, targetAddr *net.UDPAddr) {
	for {
		fmt.Print("Message to send: ")
		message := make([]byte, 4096)
		fmt.Scanln(&message)
		messageRequest := shared.CraftMessage(2, nick, string(message))
		jsonRequest, err := json.Marshal(messageRequest)
		if err != nil {
			log.Print("Error attempting to marshall: ", messageRequest, err)
			continue
		}
		UDPChannel.WriteToUDP(jsonRequest, targetAddr)
	}
}


func KeepReceiving(UDPChannel *net.UDPConn) {
	for {
		message, _, err := shared.Receive(UDPChannel)
		if err != nil {
			log.Fatal("Error while receiving", err)
		}
		fmt.Println("Incoming data from", message.Header.Peer, ":", message.Content)
	}
}

func GetCmdArguments() (string, string, string, string) {
	if len(os.Args) != 5 {
		log.Fatal("usage: ", os.Args[0], "server nick target port")
	}

	port := fmt.Sprintf(":%s", os.Args[4])
	return port, os.Args[1], os.Args[2], os.Args[3]
}

func WaitForTargetAddress(UDPChannel *net.UDPConn, nick string, target string, serverAddress *net.UDPAddr) string {
	targetAddressStr := ""
	for targetAddressStr == "" {
		targetAddressStr = LookupTarget(UDPChannel, nick, target, serverAddress)
		if targetAddressStr == "" {
			log.Println("Did not get target's address yet...")
			time.Sleep(1 * time.Second)
			continue
		}
	}
	return targetAddressStr
}

func main() {
	port, serverAddr, nick, target := GetCmdArguments()
	serverAddress, UDPChannel, err := SetupNetwork(serverAddr, port)
	if err != nil {
		log.Print("Failed to setup network", err)
	}

	err = UpdateServer(UDPChannel, nick, serverAddress)
	if err != nil {
		log.Fatal("Failed to update server directory", err)
	}

	targetAddressStr := WaitForTargetAddress(UDPChannel, nick, target, serverAddress)
	targetAddress, err := net.ResolveUDPAddr("udp4", targetAddressStr)
	if err != nil {
		log.Fatal("Resolve target address failed.", err)
	}

	go KeepReceiving(UDPChannel)
	KeepSending(UDPChannel, nick, targetAddress)
}


