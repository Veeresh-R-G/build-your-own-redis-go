package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Redis struct {
	EntryTime   map[string]time.Time
	ExpriryTime map[string]int
	Store       map[string]string
}

type ReplicationConfig struct {
	workerLock sync.Mutex
	workers    []net.Conn
}

var (
	replConf = ReplicationConfig{
		workers:    []net.Conn{},
		workerLock: sync.Mutex{},
	}
)

var (
	store = make(map[string]string)
)

type InstanceConfig struct {
	MasterId       string
	MasterPort     string
	ReplicaOf      string
	port           int
	IsSlave        bool
	IsActive       bool
	RDBFileContent []byte
}

var RDBContent string = "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
var Replicate = false

func toBulkString(value string) []byte {
	if len(value) == 0 {
		return []byte("$-1\r\n")
	}
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
}

func toSimpleString(value string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", value))
}

func toArray(values []string) []byte {
	str := ""
	for _, value := range values {
		str = fmt.Sprintf("%s%s", str, toBulkString(value))
	}
	return []byte(fmt.Sprintf("*%d\r\n%s", len(values), str))
}

func ReplicateSet(key, value string) {
	fmt.Printf("Replicating : %v : %v\n", key, value)
	fmt.Println("worker:", replConf.workers)
	for _, worker := range replConf.workers {
		_, err := worker.Write(toArray([]string{"SET", key, value}))
		if err != nil {
			log.Fatalf("Error while writing to worker : %v", err)
		}
	}
}

var isWorker bool = false

func Dispatcher(conn net.Conn, master bool) {

	buff := make([]byte, 1024)
	EntryTime := make(map[string]time.Time)
	ExpriryTime := make(map[string]int)

	response := []byte("+PONG\r\n")
	// store := make(map[string]string )
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
			ReplicateSet(key, value)
			fmt.Println(store)
		} else if cmd == "get" {
			//get
			key := tokens[4]
			fmt.Printf("GET ==> %s\n", tokens[4])
			fmt.Println(store[tokens[4]])
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
				temp := fmt.Sprintf("$%d\r\n%s\r\n", 87, "role:master\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nmaster_repl_offset:0")
				response = []byte(temp)
			} else {
				response = []byte("$10\r\nrole:slave\r\n")
			}
		} else if cmd == "replconf" {
			fmt.Println("REPLCONF hola hola")
			fmt.Println("Printing Tokens : ", tokens)
			if len(tokens) >= 5 && tokens[4] == "GETACK" {
				_, err := conn.Write(toArray([]string{"REPLCONF", "ACK", "0"}))
				if err != nil {
					fmt.Println("Error in GETACK : ", err)
				}
			}
			response = []byte("+OK\r\n")
		} else if cmd == "psync" {
			fmt.Println("PSYNC")
			response = []byte("+FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0\r\n")
			_, err = conn.Write(response)
			if err != nil {
				fmt.Printf("Error in writing response : %v\n", err)
				break
			}

			decode, err := base64.StdEncoding.DecodeString(RDBContent)
			if err != nil {
				log.Fatalln("Error while converting: ", err)
			}

			//So whenever we send a PSYNC to a worker, we append that conn object associated
			//with that worker in this array so that we know which conn object belongs to which worker
			//I'm feeling like GOD

			replConf.workerLock.Lock()
			defer replConf.workerLock.Unlock()
			replConf.workers = append(replConf.workers, conn)

			response = []byte(fmt.Sprintf("$%d\r\n%s", len(decode), decode))
			_, err = conn.Write(response)
			if err != nil {
				log.Fatalf("Error in writing response : %v\n", err)
				return
			}

			return

		}

		_, err = conn.Write(response)
		if err != nil {
			fmt.Printf("Error in writing response : %v\n", err)
			break
		}

	}
}

func sendPing(masterAddr, masterPort string) net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", masterAddr, masterPort))

	if err != nil {
		fmt.Printf("Error while creating connection object %v\n", err)
	}

	// defer conn.Close()
	buff := make([]byte, 1024)

	if _, err := conn.Write([]byte("*1\r\n$4\r\nping\r\n")); err != nil {
		log.Println("Error sending PING command to", masterAddr, ":", err)
	}
	if _, err = conn.Read(buff); err != nil {
		log.Fatalln("Not receiving response from master node")
	}

	if _, err = conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n")); err != nil {
		log.Fatalln("First Message not sent : ", err)
	}

	if _, err = conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n")); err != nil {
		log.Fatalln("Second Message not sent : ", err)
	}

	n, err := conn.Read(buff)
	if err != nil {
		log.Fatalln("Not receiving response from master node for REPLCONF ")
		os.Exit(1)
	}
	fmt.Printf("Ack - 2 from Master Node : %s", string(buff[:n]))

	n, err = conn.Read(buff)
	if err != nil {
		log.Fatalln("Not receiving ack for PSYNC from master node")
		os.Exit(1)
	}
	fmt.Printf("Ack - 3 from Master Node : %s", string(buff[:n]))

	if _, err := conn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n")); err != nil {
		log.Fatalln("Error writing PSYNC : ", err)
		os.Exit(1)
	}

	n, err = conn.Read(buff)
	if err != nil {
		log.Fatalln("Error while reading FULLRESYNC", err)
		os.Exit(1)
	}
	fmt.Printf("Ack - PSYNC from Master Node : %s\n", string(buff[:n]))

	return conn
}

func write(c net.Conn, reply []byte) error {
	_, err := c.Write(reply)
	if err != nil {
		fmt.Println("Error writing:", err.Error())
	}
	return err
}

func handleRequest(conn net.Conn, b []byte) error {
	commands := strings.Split(string(b), "\r\n")

	if len(commands) == 0 {
		conn.Write([]byte("+Empty Command \r\n"))
	}

	if len(commands) >= 1 {
		_, err := strconv.Atoi(commands[1][1:])
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	// cmd := strings.ToLower(commands[2])
	fmt.Println("=======")
	fmt.Println(commands)
	fmt.Println("=======")
	command := strings.ToLower(commands[2])
	fmt.Println("Here in Handle Request")

	switch command {
	case "ping":
		{
			return write(conn, toSimpleString("PONG"))
		}
	case "echo":
		{
			return write(conn, toBulkString(commands[4]))
		}
	case "replconf":
		{
			if commands[4] == "GETACK" {
				return write(conn, toArray([]string{"REPLCONF", "ACK", "0"}))
			}
			return write(conn, toSimpleString("OK"))
		}
	case "psync":
		{
			response := []byte("+FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0\r\n")
			_, err := conn.Write(response)
			if err != nil {
				fmt.Printf("Error in writing response : %v\n", err)
				return err

			}

			decode, err := base64.StdEncoding.DecodeString(RDBContent)
			if err != nil {
				log.Fatalln("Error while converting: ", err)
			}

			replConf.workerLock.Lock()
			defer replConf.workerLock.Unlock()
			replConf.workers = append(replConf.workers, conn)

			response = []byte(fmt.Sprintf("$%d\r\n%s", len(decode), decode))
			_, err = conn.Write(response)
			if err != nil {
				log.Fatalf("Error in writing response : %v\n", err)
				return err
			}
		}
	case "set":
		{
			key := commands[4]
			value := commands[6]
			fmt.Printf("SET ==> %s\n", commands[4])

			response := []byte("+OK\r\n")
			store[key] = value
			if isWorker {
				ReplicateSet(key, value)
			}
			_, err := conn.Write(response)
			if err != nil {
				fmt.Printf("Error in writing response : %v\n", err)
				break
			}
			fmt.Println(store)
		}
	case "get":
		{
			response := make([]byte, 1024)
			if val, ok := store[commands[4]]; !ok {
				response = []byte("$-1\r\n")
			} else {
				response = []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
			}

			_, err := conn.Write(response)
			if err != nil {
				fmt.Printf("Error in writing response : %v\n", err)
				break
			}
		}
	}

	return nil

}

func handleConnectionToMaster(conn net.Conn) {
	defer conn.Close()

	for {
		buffCmds := make([]byte, 5000)
		_, err := conn.Read(buffCmds)
		if err != nil {
			fmt.Printf("Error while receiving the commands : %v\n", err)
			break
		}

		arrayCmds := bytes.Split(buffCmds, []byte{'*'})[1:]
		for _, arr := range arrayCmds {
			arr = bytes.Trim(arr, "\x00")
			if len(arr) == 0 {
				return
			}
			arr = append([]byte("*"), arr...)
			handleRequest(conn, arr)
		}
		// fmt.Printf("The commands : %v\n", bytes.Split(buffCmds, []byte{'*'})[1:])

	}

}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	args := os.Args
	fmt.Println(args)
	port := "6379"

	if len(args) > 1 && (args[1] == "--port" || args[1] == "-p") {
		port = args[2]
	}

	master := true
	if len(args) > 3 && args[3] == "--replicaof" {
		masterConn := sendPing(args[4], args[5])
		isWorker = true
		defer masterConn.Close()
		go handleConnectionToMaster(masterConn)
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
		defer conn.Close()

		go Dispatcher(conn, master)
	}
}
