// Setups a server that receives an encoded message and print it

package main

import (
	//"fmt"
	"p2p/p2pb"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"strings"
	"strconv"
	"p2p/actors"
	"time"
)

func main() {

	// Number of processes
	n := 1000
	mainCh := make(chan p2pb.TestMessage, n)
	
	// Adresses - from ports 20000 to 20000 + n-1
	// Main adress is 8999
	addrs := make([]string, n)
	
	for i := 0; i < n; i++ {
		addr := strings.Builder{}
		addr.WriteString("127.0.0.1:")
		addr.WriteString(strconv.Itoa(20000+i))
		
		addrs[i] = addr.String()
	}
	
	mainAddr := "127.0.0.1:8999"
	
	// Creating processes
	for i := 0; i < n; i++ {
		go actors.New(i, addrs, mainAddr)
	}
	
	log.Println("Main: processes created, adresses - ", addrs)
	
	// Waiting for setup
	time.Sleep(10000*time.Millisecond)
	
	log.Println("\nMain: Ping Pong game begins!")
	
	conn := listenToPort(mainAddr)
	go persistentList(conn, mainCh)
	
	// Sending init message to processes
	for i := 0; i < n; i++ {
		go startMsg(i, addrs, conn)
	}
	
	for i := 0; i < n; i++ {
		<-mainCh
		log.Println("Main: finished ", i+1)
	}

}

func listenToPort(mainAddr string) *net.UDPConn {

	address := strings.Split(mainAddr,":")
	//log.Println(address[0])
	//log.Println(address[1])
	port, err := strconv.Atoi(address[1])
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(address[0]),
        }

	pConn, err := net.ListenUDP("udp", &addr)
	
	if err != nil {
		log.Fatalf("Main: Listening failed. %d\n", err)
	}
	log.Printf("Main: Listening to %s.\n", mainAddr)
	
	return pConn
	
}

func persistentList(pConn *net.UDPConn, pMainCh chan p2pb.TestMessage){

	for {
		// Reading size byte first
		buf := make([]byte, 1024)
		size, _, err := pConn.ReadFromUDP(buf)
		
		if err != nil {
			log.Printf("Main: Failed when reading an incoming message. %d\n", err)
			pConn.Close()
			return
		}
		
		msg := &p2pb.TestMessage{}
		err = proto.Unmarshal(buf[:size], msg)
		
		log.Printf("Main: Received a finishing message from P%s.\n", msg.Sender)
		
		pMainCh <- *msg
	}

}

func startMsg(p int, addrs []string, pConn *net.UDPConn) {
	
	// Sending Greet message
	msg2 := &p2pb.TestMessage{Sender: strconv.Itoa(-1), Receiver: strconv.Itoa(p), Text: "GREET"}
	eMsg, err := proto.Marshal(msg2)
	
	if err != nil {
		log.Printf("Main: Error when encoding a message. %d\n", err)
		pConn.Close()
		return
	}
	
	address := strings.Split(addrs[p],":")
	port, err := strconv.Atoi(address[1])
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(address[0]),
        }
        //log.Printf("Address %v \n", addr)
	
	_, err = pConn.WriteToUDP(eMsg, &addr)
	if err != nil {
		log.Printf("Response err %v", err)
	}
}

