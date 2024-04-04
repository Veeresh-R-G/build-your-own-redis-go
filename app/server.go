package main

import (
	"fmt"
	"net"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here !")

	// Uncomment this block to pass the first stage
	//
	listener, err := net.Listen("tcp", "localhost:6379")
	if err != nil {
		fmt.Printf("Error = %s", err)
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		fmt.Printf("Error = %s", err)
	}

	response := []byte("+PONG\r\n")
	for {

		buff := make([]byte, 1024)

		_, err := conn.Read(buff)
		if err != nil {
			fmt.Printf("Error Reading the Request : %v", err)
			break
		}

		_, err = conn.Write(response)
		if err != nil {
			fmt.Printf("Error = %s", err)
			break
		}

	}

	conn.Close()

}
