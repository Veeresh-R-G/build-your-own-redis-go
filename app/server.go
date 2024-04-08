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

type Redis struct {
	EntryTime   map[string]time.Time
	ExpriryTime map[string]int
	store       map[string]string
}

func Dispatcher(conn net.Conn, master bool) {
	defer conn.Close()
	buff := make([]byte, 1024)
	EntryTime := make(map[string]time.Time)
	ExpriryTime := make(map[string]int)

	response := []byte("+PONG\r\n")
	store := make(map[string]string)
	for {

		n, err := conn.Read(buff)
		if err != nil {
			fmt.Printf("Error Reading the Request : %v\n", err)
			break
		}

		tokens := strings.Split(string(buff[:n]), "\r\n")
		tokens = tokens[:len(tokens)-1]
		fmt.Println("-------")
		fmt.Println(tokens)
		fmt.Println("-------")
		cmd := strings.ToLower(tokens[2])

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
			store[key] = value

			fmt.Println(store)
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
					if val, ok := store[tokens[4]]; !ok {
						response = []byte("$-1\r\n")
					} else {
						response = []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
					}
				}
			} else {
				if val, ok := store[tokens[4]]; !ok {
					response = []byte("$-1\r\n")
				} else {
					response = []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
				}
			}

		} else if cmd == "info" && strings.ToLower(tokens[4]) == "replication" {
			fmt.Println("INFO =====>")
			if master {
				fmt.Println("Here in master")

				temp := fmt.Sprintf("$%d\r\n%s\r\n", 87, "role:master\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nmaster_repl_offset:0")

				response = []byte(temp)
			} else {
				response = []byte("$10\r\nrole:slave\r\n")
			}
		}

		_, err = conn.Write(response)
		if err != nil {
			fmt.Printf("Error in writing response : %v\n", err)
			break
		}

	}
}

func sendPing(masterAddr, masterPort string) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", masterAddr, masterPort))

	if err != nil {
		fmt.Printf("Error while creating connection object %v\n", err)
	}

	defer conn.Close()
	buff := make([]byte, 1024)

	if _, err := conn.Write([]byte("*1\r\n$4\r\nping\r\n")); err != nil {
		log.Println("Error sending PING command to", masterAddr, ":", err)
		return
	}
	if _, err = conn.Read(buff); err != nil {
		log.Fatalln("Not receiving response from master node")
	}

	if _, err = conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n")); err != nil {
		log.Fatalln("First Message not sent : ", err)
		return
	}

	if _, err = conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n")); err != nil {
		log.Fatalln("Second Message not sent : ", err)
	}

	if _, err = conn.Read(buff); err != nil {
		log.Fatalln("Not receiving response from master node")
		os.Exit(1)
	}

}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here !")

	args := os.Args
	fmt.Println(args)
	port := "6379"

	if len(args) > 1 && (args[1] == "--port" || args[1] == "-p") {
		port = args[2]
	}

	master := true
	if len(args) > 3 && args[3] == "--replicaof" {
		sendPing(args[4], args[5])
		master = false
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		fmt.Printf("Error in initiating tcp connection = %s\n", err)
	}
	defer listener.Close()

	for {

		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error in creating connection object %s\n", err)
			break
		}

		fmt.Println("Here I am")
		go Dispatcher(conn, master)
	}

}
