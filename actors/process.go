// 

package actors

import (
	"fmt"
	"p2p/p2pb"
	"google.golang.org/protobuf/proto"
	"log"
	"strconv"
	"net"
	//"time"
)


// Main atributes of the process
// addrs[id] is the process' own adress
type state struct {
	id  int // Process' own id
	addrs map[int]net.UDPAddr // Internet adresses of peers
	mainAddr string // Addres of main process
	myChan chan []byte // Received messages from mailBox
	mailBox peerConn // Mail box object
	counter int // Number of PONGS received
}


// Getters
func (s *state) Id() int {
	return s.id
}

func (s *state) Addrs() map[int]net.UDPAddr {
	return s.addrs
}

func (s *state) MainAddr() string {
	return s.mainAddr
}

// New process creation
func New(pId int, pAddrs map[int]net.UDPAddr, pMain string) {
/*
	pId = Processes' own identifier
	pAddrs = Processes addresses
	pMain = Main application address
*/

	fmt.Printf("Created process P%d: %s\n", pId, pAddrs[pId])
	
	//time.Sleep(1000*time.Millisecond)
	
	s := state{id: pId,
		addrs: pAddrs,
		mainAddr: pMain}

	s.myChan = make(chan []byte, 100)
	s.mailBox = NewMailBox(pId, pAddrs, s.myChan)
	
	go s.mailBox.SetupConnections()
	
	//sum := 0
	for m := range s.myChan {
		//sum ++
		//log.Printf("P%d: Counter %d.\n", s.id, sum)
		s.onReceive(m)
		//log.Printf("Received message here!")
		
	}
	
}

// Message Handlers
func (s *state) onReceive(m []byte) {
	
	// Identify received bytes and unmarshal message here
	msg := &p2pb.TestMessage{}
	err := proto.Unmarshal(m, msg)
	
	if err != nil {
		log.Printf("P%d: Failed when unmarshalling message. %d\n", s.id, err)
		log.Printf("P%d: Received bytes: %d. \n", s.id, m)
		return
	}
	
	//log.Printf("P%d: Received messase of ....................... SIZE %d.\n", s.id, len(m))
	//log.Printf("P%d: Received messase    .......................  %d.\n", s.id, m)
	//log.Printf("P%d: Received messase    .......................  %d.\n", s.id, msg)

	if msg.Text == "PING" {
		s.pingHandler(*msg)
	} else if msg.Text == "PONG" {
		s.pongHandler(*msg)
	} else if msg.Text == "GREET" {
		s.greetHandler(*msg)
	}
	
}

func (s *state) greetHandler(m p2pb.TestMessage) {
	for p := range s.addrs{
		if p != s.id{
			log.Printf("P%d: Sending PING to process %d.\n", s.id, p)
			msg := &p2pb.TestMessage{Sender: strconv.Itoa(s.id), Receiver: strconv.Itoa(p), Text: "PING"}
			s.sendMsg(p, msg)
		}
	}
}

func (s *state) pingHandler(m p2pb.TestMessage) {
	log.Printf("P%d: Received PING from %s. Replying PONG.\n", s.id, m.Sender)
	msg := &p2pb.TestMessage{Sender: strconv.Itoa(s.id), Receiver: m.Sender, Text: "PONG"}
	sender, err := strconv.Atoi(m.Sender)
	
	if err != nil {
		log.Printf("P%d: Error when converting string. %d\n", s.id, err)
		return
	}
	
	s.sendMsg(sender, msg)
}

func (s *state) pongHandler(m p2pb.TestMessage) {
	log.Printf("P%d: Received PONG from %s.\n", s.id, m.Sender)
	s.counter++
	if s.counter == len(s.addrs)-1 {
		//msg := &p2pb.TestMessage{Sender: strconv.Itoa(s.id), Receiver: "-1"}
		msg := &p2pb.TestMessage{Sender: strconv.Itoa(s.id), Receiver: "-1"}
		s.sendMsg(-1, msg)
	}
}

func (s *state) sendMsg(pId int, msg *p2pb.TestMessage) {
	
	// Encode message and call sending method from mail box object
	eMsg, err := proto.Marshal(msg)
	
	if err != nil {
		log.Printf("P%d: Error when encoding a message. %d\n", s.id, err)
		return
	}
	
	s.mailBox.SendMsg(pId, eMsg)
	
}
