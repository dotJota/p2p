package actors

import (
	"log"
	"p2p/p2pb"
	"net"
	"google.golang.org/protobuf/proto"
	"time"
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
	conns []net.Conn // Connections with peers
	mainConn net.Conn // Connection with main application
}

func newMailBox(ownId int, peerAddrs []string, mainAddresses string, outChannel chan []byte) peerConn {

	pC := peerConn{ id: ownId,
		addrs: peerAddrs,
		mainAddr: mainAddresses,
		outChan: outChannel }
		
	pC.conns = make([]net.Conn, len(peerAddrs))
	
	return pC
	
}

func (pC *peerConn) SetupConnections() {
	
	go pC.connectToPeers()
	pC.waitForConnections()
	
}

// From 0 to n-1 do: dial to every >i+1 connection and listen to every <i-1 connection

func (pC *peerConn) waitForConnections() {
	
	// First process dials with everyone else
	//if pC.id == 0 {
	//	return
	//}

	log.Printf("P%d: Listening to %s.\n", pC.id, pC.addrs[pC.id])
	ls, err := net.Listen("tcp", pC.addrs[pC.id])
	
	if err != nil {
		log.Fatalf("P%d: Listening failed. %d\n", pC.id, err)
	}
	
	for {
		// Listen for incoming connection
		conn, err := ls.Accept()
		
		if err != nil {
			log.Printf("P%d: Connection failed. %d\n", pC.id, err)
			conn.Close()
		} else {
			go pC.handleConnection(conn)
		}
		
	}
	
}

func (pC *peerConn) handleConnection(pConn net.Conn) {

	// Waiting for sender's id
	buf := make([]byte, 1024)
	l, err := pConn.Read(buf)
	
	msg := &p2pb.ConnMessage{}
	err = proto.Unmarshal(buf[:l], msg)
	
	if err != nil {
		log.Printf("P%d: Unable to unmarchal id message. %d\n", pC.id, err)
		pConn.Close()
		return
	}
	
	if msg.Sender == -1 {
		pC.mainConn = pConn
		log.Printf("P%d: Connection established with Main.\n", pC.id)
		
	} else {
		pC.conns[msg.Sender] = pConn
		log.Printf("P%d: Connection established with P%d.\n", pC.id, msg.Sender)
	}
	
	pConn.Write([]byte{1})
	
	pC.persistentList(pConn)
	
}

// Logic for establishing connections
func (pC *peerConn) connectToPeers() {
	
	time.Sleep(5000*time.Millisecond)
	
	for i := pC.id + 1; i < len(pC.addrs); i++ {
		go pC.connect(i)
	}
}

func (pC *peerConn) connect(p int) {
	
	// Establishing a connection
	conn, err := net.Dial("tcp", pC.addrs[p])
	
	if err != nil {
		log.Printf("P%d: Failed to establish a connection to P%d. %d\n", pC.id, p, err)
		conn.Close()
	}
	
	// Greet Message - Informing sender
	msg := &p2pb.ConnMessage{Sender: int64(pC.id), Offset: 1}
	eMsg, err := proto.Marshal(msg)
	
	if err != nil {
		log.Printf("P%d: Error when encoding a message. %d\n", pC.id, err)
		conn.Close()
		return
	}
	
	conn.Write(eMsg)
	
	// Receiving reply
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	
	if err != nil {
		log.Printf("P%d: Failed to read reply message from P%d. %d\n", pC.id, p, err)
		conn.Close()
		return
	}
	
	// Assigning correct slot to connection
	pC.conns[p] = conn
	
	pC.persistentList(conn)
	
}

func (pC *peerConn) persistentList(conn net.Conn){

	for {
		// Reading size byte first
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		
		buf = make([]byte, int(buf[0]))
		_, err = conn.Read(buf)
		
		
		if err != nil {
			log.Printf("P%d: Failed when reading an incoming message. %d\n", pC.id, err)
			conn.Close()
			return
		}
		
		pC.outChan <- buf
	}

}

func (pC *peerConn) SendMsg(r int, m []byte) {
	
	// Adding message size on first byte
	m_ := make([]byte, len(m)+1)
	m_[0] = byte(len(m))
	
	for i := 0; i < len(m); i++ {
		m_[i+1] = m[i]
	}
	
	if r == -1 {
		pC.mainConn.Write(m_)
	} else {
		pC.conns[r].Write(m_)
	}
	
}


