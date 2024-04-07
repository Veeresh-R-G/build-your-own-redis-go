package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func Dispatcher(conn net.Conn) {
	defer conn.Close()
	buff := make([]byte, 1024)

	EntryTime := make(map[string]time.Time)
	ExpriryTime := make(map[string]int)

	response := []byte("+PONG\r\n")
	mp := make(map[string]string)
	for {

		n, err := conn.Read(buff)
		if err != nil {
			fmt.Printf("Error Reading the Request : %v\n", err)
			break
		}

		tokens := strings.Split(string(buff[:n]), "\r\n")
		tokens = tokens[:len(tokens)-1]
		fmt.Println(tokens)
		cmd := tokens[2]

		if cmd == "echo" {
			fmt.Println("ECHO ===")
			response = []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(tokens[4]), tokens[4]))
		} else if cmd == "set" {
			key := tokens[4]
			value := tokens[6]
			fmt.Printf("SET ==> %s\n", tokens[4])

			if len(tokens) > 7 {
				//px command
				EntryTime[key] = time.Now()
				time_duration := tokens[10]
				num, err := strconv.Atoi(time_duration)
				if err != nil {
					log.Fatalln("Error While converting to a number")
				}
				ExpriryTime[key] = num
			}
			response = []byte("+OK\r\n")
			mp[key] = value

			fmt.Println(mp)
		} else if cmd == "get" {
			//get
			key := tokens[4]
			fmt.Printf("GET ==> %s\n", tokens[4])
			if val, ok := ExpriryTime[key]; ok {

				if time.Since(EntryTime[key]).Milliseconds() >= int64(val) {
					//Deadline is reached
					fmt.Println(" ---- TimeOUT ---- ")
					delete(EntryTime, key)
					delete(ExpriryTime, key)
					response = []byte("$-1\r\n")
				} else {
					fmt.Println(" ---- Not expired Key ---- ")
					if val, ok := mp[tokens[4]]; !ok {
						response = []byte("$-1\r\n")
					} else {
						response = []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
					}
				}
			} else {
				if val, ok := mp[tokens[4]]; !ok {
					response = []byte("$-1\r\n")
				} else {
					response = []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
				}
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

	args := os.Args
	fmt.Println(args)
	port := "6379"

	if len(args) > 1 && args[1] == "--port" {
		port = args[2]
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
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
