package tserver

import (
	"golangServer/config"
	"errors"
	"log"
	"net"
	"os"
	"strings"
)

type Server struct {
	Port string
}

var ClientList = []*Client{}

var UDPConn *net.UDPConn
var UDPAddr *net.UDPAddr

var ConnMap = make(map[uint32]net.Conn)

func StartServer(port string) Server {
	server := Server{Port: port}
	Start(server)
	return server
}

func Start(server Server) {
	servicePort := server.Port
	log.Println("Starting Tcp Server", servicePort)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", servicePort)
	checkErr(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkErr(err)

	//Start Udp Server
	uDPAddr, err := net.ResolveUDPAddr("udp", config.UDPport)
	UDPAddr = uDPAddr
	checkErr(err)
	uDPConn, err := net.ListenUDP("udp", UDPAddr)
	UDPConn = uDPConn
	checkErr(err)

	go func() {
		for {
			HandleClientUdp(UDPConn)
		}
	}()
	var connid uint32
	for {
		// Here when listen an conn to New a Client !
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept failed, err:", err)
			continue
		}
		connid++
		uniid := connid
		ConnMap[uniid] = conn
		go handleClient(conn, uniid)
	}
}

func handleClient(conn net.Conn, uniid uint32) {
	defer conn.Close()

	client := InstanceClient(conn, uniid)
	ClientList, _ = AddClient(client)
	buf := make([]byte, 1024)
	for {
		cnt, err := conn.Read(buf)
		if checkNetErr(err, conn) {
			log.Println(" conn.Read error", err)
			RemoveClient(client)
			break
		}
		err2 := handBuffer(client, buf[BufferHeadLength:cnt])
		if checkNetErr(err2, conn) {
			log.Println("handBuffer error", err2)
			RemoveClient(client)
			break
		}
	}
}

func AddClient(client *Client) ([]*Client, error) {
	ClientList = append(ClientList, client)
	log.Println("Add client to Server =>", client.Addr, "  ClientCount =>", len(ClientList))

	if len(ClientList) > 3000 {
		os.Exit(1)
	}
	return ClientList, nil
}

func GetClient(clientName string) (*Client, error) {
	for i := 0; i < len(ClientList); i++ {
		if ClientList[i] == nil {
			continue
		}
		if ClientList[i].Username == "" {
			log.Println("Had client Username = nil 👉", clientName)
			ClientList[i].Username = clientName
			return ClientList[i], nil
			// continue
		}
		if strings.Compare(ClientList[i].Username, clientName) == 0 {
			return ClientList[i], nil
		}
	}

	log.Println("Had not client 👉", clientName)
	return nil, errors.New("no client")
}

func RemoveClient(client *Client) {
	ClientList = RemoveC(ClientList, client)
	_, ok := ConnMap[client.Uniid]
	if ok {
		delete(ConnMap, client.Uniid)
	}
	log.Println("Rmv client from Server =>", client.Addr, "Uniid :=>", client.Uniid, "  ClientCount =>", len(ClientList))
}

func checkErr(err error) {
	if err != nil {
		log.Printf("Fatal error: %v", err.Error())
		// os.Exit(1)
	}
}

func checkNetErr(err error, conn net.Conn) bool {
	if err != nil {
		log.Println(conn.RemoteAddr())
		log.Printf("Fatal error: %v", err.Error())
		return true
	}
	return false
}
