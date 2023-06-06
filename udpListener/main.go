// Setups a server that receives an encoded message and print it

package main

import (
	"p2p/actors"
	"log"
	"net"
	"time"
	"flag"
)

var (

	baseIpAddress = flag.String("ip", "10.0.0.1", "Base IP address to generate all peer's IPs")
	
	port = flag.Int("port", 5001, "Base port to generate all peer's port")
	
	wait = flag.Int("wait", 60, "Listening expiration time in seconds")

)

func main() {

	numberOfPeers := 1
	processIndex := 0
	
	boxChan := make(chan []byte, 1000)
	flag.Parse()
	
	addresses := make(map[int]net.UDPAddr)
	for i:=0; i < numberOfPeers; i++ {
		baseIp := net.UDPAddr {
			Port: *port,
			IP: net.ParseIP(*baseIpAddress),
		}
		baseIp.Port = baseIp.Port + i
		baseIp.IP[15] = byte(i) + baseIp.IP[15]
		addresses[i] = baseIp
	}
	
	log.Println("Hello!")
	
	mail := actors.NewMailBox(processIndex, addresses, boxChan)
	mail.SetupConnections()
	
	waitChan := make(chan bool)
	go waitDiscovery(boxChan)
	
	select{
		case <-time.After(time.Duration(*wait * 1000000000)):
		case <- waitChan:
	}
	
	log.Printf("Finished!")
	
}

func waitDiscovery(pChan chan []byte) {
	for bytes := range pChan{
		log.Printf("Received: %s\n", bytes)
	}
}
