package actors

import (
	"log"
	//"p2p/p2pb"
	"net"
	//"google.golang.org/protobuf/proto"
	"time"
	//"strconv"
	//"strings"
	"bufio"
	"math"
)

/*

The mailBox object abstracts communication between peers and a main (setup) application.
It is agnostic of the type of messages communicated: it treats them as slices of bytes,
which can be in an encoded form.

Receives:
	
	ownId
	peerAdresses, mainAdress
	outChannel
	
Establishes:
	connections with peers
	connection with main
	
Exports:
	SetupConnections(ownId int, peerAdresses []string, mainAdress string, outChannel chan []byte)
	SendMsg(msg []byte, dest int)

Received messages are sent to outChannel, the calling process should listen to this channel in order to treat them.

*/

// addrs[id] is the process' own adress
type peerConn struct {
	id  int // Process' own id
	addrs map[int]net.UDPAddr // Internet adresses of peers
	outChan chan []byte // Send messages to process
	conn *net.UDPConn
}

func (pC *peerConn) GetConn() *net.UDPConn {
	return pC.conn
}

func NewMailBox(ownId int, peerAddrs map[int]net.UDPAddr, outChannel chan []byte) peerConn {

	pC := peerConn{ id: ownId,
		addrs: peerAddrs,
		outChan: outChannel }
		
	//pC.conns = make([]net.Conn, len(peerAddrs))
	
	return pC
	
}

func (pC *peerConn) SetupConnections() {

	ip := pC.addrs[pC.id]
	conn, err := net.ListenUDP("udp", &ip)
	
	if err != nil {
		log.Fatalf("P%d: Listening failed. %s\n", pC.id, err)
	}
	log.Printf("P%d: Listening to %s.\n", pC.id, &ip)
	
	pC.conn = conn
	pC.conn.SetReadBuffer(int(math.Pow(2,32)-1))
	//pC.conn.SetWriteBuffer(int(math.Pow(2,32)-1))
	go pC.persistentList()
	
}


func (pC *peerConn) persistentList(){
	
	reader := bufio.NewReaderSize(pC.conn,100000)
	time.Sleep(10000000000)

	for {
		buf := make([]byte, 20)
		size, err := reader.Read(buf)
		
		if err != nil {
			log.Printf("P%d: Failed when reading an incoming message. %d\n", pC.id, err)
			return
		}
		
		//log.Printf("Size: %d \n",size)
		
		go func(){pC.outChan <- buf[:size]}()
	}

}


func (pC *peerConn) SendMsg(r int, m []byte) {
	
	go func(pR int, pM []byte){
		
		ip := pC.addrs[pR]
		_, err := pC.conn.WriteToUDP(pM, &ip)
		if err != nil {
			log.Printf("Response err %v", err)
		}
	
	}(r, m)
}


