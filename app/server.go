package main

import (
	"fmt"
	"net"
	"strings"
)

func Dispatcher(conn net.Conn) {
	defer conn.Close()
	buff := make([]byte, 1024)

	response := []byte("+PONG\r\n")
	mp := make(map[string]string)
	for {

		n, err := conn.Read(buff)
		if err != nil {
			fmt.Printf("Error Reading the Request : %v\n", err)
			break
		}

		tokens := strings.Split(string(buff[:n]), "\r\n")
		fmt.Println(tokens)
		cmd := tokens[2]

		if cmd == "echo" {
			response = []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(tokens[4]), tokens[4]))
		} else if cmd == "set" {
			response = []byte("+OK\r\n")
			mp[tokens[4]] = tokens[6]
			fmt.Println(mp)
		} else if cmd == "get" {
			//get
			fmt.Printf("GET ==> %s\n", tokens[4])
			if val, ok := mp[tokens[4]]; !ok {
				response = []byte("$-1\r\n")
			} else {
				response = []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
			}
		}

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
			break
		}

		go Dispatcher(conn)
	}

}
