// A client that sends a message to a server

package main

import (
	"fmt"
	"p2p/p2pb"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"time"
)

func main() {

	msg := &p2pb.TestMessage{Sender: "Me", Text: "Just a test message."}
	
	// Encoding message
	eMsg, err := proto.Marshal(msg)
	if err != nil {
		log.Fatalln("Error when encoding the message, ", err)
	}
	
	// Establishing a connection
	addr := "localhost:8888"
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to connect to a server, ", err)
	}
	
	fmt.Println("Sending message now...")
	
	conn.Write(eMsg)
	
	fmt.Println("Message Sent.")
	
	// Receiving reply
	buf := make([]byte, 1024)
	l,err := conn.Read(buf)
	if err != nil {
		log.Fatalln("Failed to read reply message, ", err)
	}
	
	fmt.Println("Received: ", string(buf[:l]))
	
	time.Sleep(10000*time.Millisecond)
	
	// Second message
	fmt.Println("Sending second message now...")
	
	conn.Write(eMsg)
	
	fmt.Println("Second message Sent.")
	
	// Receiving reply
	buf = make([]byte, 1024)
	l,err = conn.Read(buf)
	if err != nil {
		log.Fatalln("Failed to read reply message, ", err)
	}
	
	fmt.Println("Received: ", string(buf[:l]))

}
