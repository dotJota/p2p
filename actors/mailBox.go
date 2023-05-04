package actors

import (
	"log"
	//"p2p/p2pb"
	"net"
	//"google.golang.org/protobuf/proto"
	//"time"
	"strconv"
	"strings"
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
	addrs []string // Internet adresses of peers
	mainAddr string // Addres of main process
	outChan chan []byte // Send messages to process
	conn *net.UDPConn
}

func newMailBox(ownId int, peerAddrs []string, mainAddresses string, outChannel chan []byte) peerConn {

	pC := peerConn{ id: ownId,
		addrs: peerAddrs,
		mainAddr: mainAddresses,
		outChan: outChannel }
		
	//pC.conns = make([]net.Conn, len(peerAddrs))
	
	return pC
	
}

func (pC *peerConn) SetupConnections() {
	pC.waitForConnections()
}


func (pC *peerConn) waitForConnections() {
	
	address := strings.Split(pC.addrs[pC.id],":")
	port, err := strconv.Atoi(address[1])
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(address[0]),
        }

	conn, err := net.ListenUDP("udp", &addr)
	
	if err != nil {
		log.Fatalf("P%d: Listening failed. %s\n", pC.id, err)
	}
	log.Printf("P%d: Listening to %s.\n", pC.id, pC.addrs[pC.id])
	
	pC.conn = conn
	go pC.persistentList()
	
}


func (pC *peerConn) persistentList(){

	for {
		buf := make([]byte, 128)
		size, _, err := pC.conn.ReadFromUDP(buf)
		
		if err != nil {
			log.Printf("P%d: Failed when reading an incoming message. %d\n", pC.id, err)
			pC.conn.Close()
			return
		}
		go func(){pC.outChan <- buf[:size]}()
	}

}


func (pC *peerConn) SendMsg(r int, m []byte) {
	
	go func(pR int, pM []byte){
		
		address := make([]string,2)
		
		if r == -1 {
			address = strings.Split(pC.mainAddr,":")
		} else {
			address = strings.Split(pC.addrs[pR],":")
		}
		port, err := strconv.Atoi(address[1])
		addr := net.UDPAddr{
			Port: port,
			IP:   net.ParseIP(address[0]),
		}
		
		_, err = pC.conn.WriteToUDP(pM, &addr)
		if err != nil {
			log.Printf("Response err %v", err)
		}
	
	}(r, m)
}


