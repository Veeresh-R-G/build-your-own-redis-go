package main

import (
	"fmt"
	"net"
)

func Dispatcher(conn net.Conn) {
	defer conn.Close()
	buff := make([]byte, 1024)
	response := []byte("+PONG\r\n")
	for {

		n, err := conn.Read(buff)
		if err != nil {
			fmt.Printf("Error Reading the Request : %v\n", err)
			break
		}
		fmt.Printf("Buffer = %v\n", buff[:n])
		_, err = conn.Write(response)
		if err != nil {
			fmt.Printf("Error in writing response : %v\n", err)
			break
		}
	}

}
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here !")

	// Uncomment this block to pass the first stage
	//
	listener, err := net.Listen("tcp", "localhost:6379")
	if err != nil {
		fmt.Printf("Error in initiating tcp connection = %s\n", err)
	}
	defer listener.Close()

	for {

		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error in creating conn object %s\n", err)
		}

		go Dispatcher(conn)
	}

}
