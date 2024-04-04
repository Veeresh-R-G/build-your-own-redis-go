package main

import (
	"fmt"
	"net"
	// Uncomment this block to pass the first stage
	// "net"
	// "os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		fmt.Printf("Error = %s", err)
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		fmt.Printf("Error = %s", err)
	}

	defer conn.Close()
	response := []byte("+PONG\r\n")

	n, err := conn.Write(response)
	if err != nil {
		fmt.Printf("Error = %s", err)
	}

	fmt.Printf("n = %d", n)

}
