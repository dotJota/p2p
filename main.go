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
	n := 30
	mainCh := make(chan p2pb.TestMessage, n)
	
	// Adresses - from ports 20000 to 20000 + n-1
	// Main adress is 8999
	addrs := make([]string, n)
	
	for i := 0; i < n; i++ {
		addr := strings.Builder{}
		addr.WriteString("localhost:")
		addr.WriteString(strconv.Itoa(20000+i))
		
		addrs[i] = addr.String()
	}
	
	mainAddr := "localhost:8999"
	
	// Creating processes
	for i := 0; i < n; i++ {
		go actors.New(i, addrs, mainAddr)
	}
	
	log.Println("Main: processes created, adresses - ", addrs)
	
	// Waiting for setup
	time.Sleep(10000*time.Millisecond)
	
	log.Println("\nMain: Ping Pong game begins!")
	
	// Connecting main with processes
	for i := 0; i < n; i++ {
		go connect(i, addrs, mainCh)
	}
	
	for i := 0; i < n; i++ {
		log.Println("Main: finished ", i+1)
		<-mainCh
	}

}

func connect(p int, addrs []string, mainCh chan p2pb.TestMessage) {

	// Establishing a connection
	conn, err := net.Dial("tcp", addrs[p])
	
	if err != nil {
		log.Printf("Main: Failed to establish a connection to P%d. %d\n", p, err)
		conn.Close()
	}
	
	// Main id = -1
	msg := &p2pb.ConnMessage{Sender: -1, Offset: 1}
	eMsg, err := proto.Marshal(msg)
	
	if err != nil {
		log.Printf("Main: Error when encoding a message. %d\n", err)
		conn.Close()
		return
	}
	
	conn.Write(eMsg)
	
	// Receiving reply
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	
	if err != nil {
		log.Printf("Main: Failed to read reply message from P%d. %d\n", p, err)
		conn.Close()
		return
	}
	
	//log.Printf("Main: Connection establish with P%d: %s\n", p, string(buf[:l]))
	
	// Sending Greet message
	msg2 := &p2pb.TestMessage{Sender: strconv.Itoa(-1), Receiver: strconv.Itoa(p), Text: "GREET"}
	eMsg, err = proto.Marshal(msg2)
	
	if err != nil {
		log.Printf("Main: Error when encoding a message. %d\n", err)
		conn.Close()
		return
	}
	
	// Adding message size on first byte
	m_ := make([]byte, len(eMsg)+1)
	m_[0] = byte(len(eMsg))
	
	for i := 0; i < len(eMsg); i++ {
		m_[i+1] = eMsg[i]
	}
	
	conn.Write(m_)
	
	//log.Printf("Main: Greet message sent to P%d.\n", p)
	
	for {	
		// Reading size byte first
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		
		buf = make([]byte, int(buf[0]))
		_, err = conn.Read(buf)
		
		
		if err != nil {
			log.Printf("Main: Failed when reading an incoming message. %d\n", err)
			conn.Close()
			return
		}
		
		msg := &p2pb.TestMessage{}
		err = proto.Unmarshal(buf, msg)
		
		log.Printf("Main: Received a finishing message from P%s.\n", msg.Sender)
		
		mainCh <- *msg
	}
	
}

