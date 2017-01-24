// Implementation of a MultiEchoServer. Students should write their code in this file.

package p0

import (
	"net"
	"fmt"
	"bufio"
)

type echoClient struct {
	connection net.Conn
	channel chan string
}

type multiEchoServer struct {
	// TODO: implement this!
	host string
	echoClientChannel chan map[int] *echoClient
	listener net.Listener
	stopChannel chan bool
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	// TODO: implement this!
	server := &multiEchoServer{
		host: "localhost",
		echoClientChannel: make(chan map[int] *echoClient, 1),
		stopChannel: make(chan bool),
	}
	return MultiEchoServer(server)
}

func (mes *multiEchoServer) Start(port int) error {
	// TODO: implement this!
	listener, err := net.Listen("tcp", fmt.Sprint(":%d", port))
	mes.listener = listener
	if err != nil{
		fmt.Println("Error on listen : ", err)
		return err
	}
	fmt.Printf("Server running on %s : %d", mes.host, port)

	// serve or accept connection. (MultiEchoServer should)
	go mes.serve()
	return nil
}

func (mes *multiEchoServer) Close() {
	// TODO: implement this!
	for _, client := range <-mes.echoClientChannel{
		client.connection.Close()
	}
	mes.listener.Close()
	close(mes.stopChannel)
}

func (mes *multiEchoServer) Count() int {
	// TODO: implement this!
	clients := <-mes.echoClientChannel
	count := len(clients)
	mes.echoClientChannel <- clients
	return count
}

func (mes *multiEchoServer) serve()  {
	messageChannel := make(chan string)
	go broadcastMessage(messageChannel, mes.echoClientChannel)

	clients := make(map[int]*echoClient)
	mes.echoClientChannel <- clients
	i:= 0
	for{
		connection, err := mes.listener.Accept()
		if err!=nil{
			select {
			case <- mes.stopChannel:
				return
			default:
			}
			fmt.Println("Error on Accept : ", err)
			continue
		}
		go handleConnection(i, connection, mes.echoClientChannel, messageChannel)
		i++
	}
}

func handleConnection(clientNumber int, connection net.Conn, echoClientChannel chan map[int] *echoClient, messageChannel chan <-string)  {
	fmt.Printf("Client : %d : %v <-> %v \n", clientNumber, connection.LocalAddr(), connection.RemoteAddr())

	EchoClientPtr := &echoClient{
		connection: connection,
		channel: make(chan string, 100),
	}

	clients := <- echoClientChannel
	clients[clientNumber] = EchoClientPtr
	echoClientChannel <- clients

	go echo(EchoClientPtr)
	defer EchoClientPtr.connection.Close()

	rb := bufio.NewReader(connection)
	for{
		message, e := rb.ReadString('\n')
		if e!=nil{
			break
		}
		messageChannel <- message
	}

	clients <- echoClientChannel
	delete(clients, clientNumber)
	echoClientChannel <- clients
	fmt.Printf("%d : Closed \n", clientNumber)
}


func echo(client *echoClient){
	for{
		message := <-client.channel
		_, err := client.connection.Write([]byte(message))
		if err!=nil{
			break
		}
	}
}

func broadcastMessage(messageChannel <-chan string, echoClientChannel chan map[int] *echoClient)  {
	for{
		message := <-messageChannel
		clients := <-echoClientChannel
		for _, echoClient := range clients{
			select {
			case echoClient.channel <- message:
			default:
			}
		}
		echoClientChannel <- clients
	}
}

