// Setups a server that receives an encoded message and print it

package main

import (
	"fmt"
	"p2p/p2pb"
	"p2p/actors"
	"google.golang.org/protobuf/proto"
	//"log"
	"net"
	//"strings"
	//"strconv"
	"time"
	"flag"
)

var (
	numberOfPeers = flag.Int("n", 3, "Number of peers")

	processIndex = flag.Int("i", 0, "Peer's id")

	baseIpAddress = flag.String("base_ip", "10.0.0.1", "Base IP address to generate all peer's IPs")
	
	port = flag.Int("base_port", 5001, "Base port to generate all peer's port")

)

func main() {

	boxChan := make(chan []byte, 1000)
	flag.Parse()
	
	addresses := make(map[int]net.UDPAddr)
	for i:=0; i<*numberOfPeers; i++ {
		baseIp := net.UDPAddr {
			Port: *port,
			IP: net.ParseIP(*baseIpAddress),
		}
		baseIp.Port = baseIp.Port + i
		baseIp.IP[15] = byte(i+1)
		addresses[i] = baseIp
	}
	
	fmt.Println("Hello!")
	
	mail := actors.NewMailBox(*processIndex, addresses, boxChan)
	mail.SetupConnections()
	
	time.Sleep(10000000000)
	
	// Sending discovery message
	msg := &p2pb.DiscoveryMessage{Sender: int32(*processIndex)}
	eMsg, err := proto.Marshal(msg)
	if err != nil {
		fmt.Printf("Error: %s \n", err)
	}
	
	numberMsg := 10000
	for i:=0; i<*numberOfPeers; i++{
		for j:=0; j<numberMsg; j++{
			mail.SendMsg(i, eMsg)
		}
	}
	
	waitChan := make(chan bool)
	counter := 0
	go waitDiscovery(boxChan, *numberOfPeers*numberMsg, waitChan, &counter)
	select{
		case <-time.After(50000000000):
		case <- waitChan:
	}
	
	fmt.Printf("Finished discovery! Received %d messages.\n", counter)
	time.Sleep(10000000000)
	
}

func waitDiscovery(pChan chan []byte, nMsg int, outChan chan bool, pCounter *int) {
	for _ = range pChan{
		//fmt.Printf("Received: %d\n", m)
		*pCounter++
		if *pCounter == nMsg {
			break
		}
	}
	outChan <- true
}
