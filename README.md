### Stage 1

1. It was just uncommenting some code that was already there.

### Stage 2

1. Opened a TCP connection server that listens to a PING message

### Stage 3

1. Learned about Redis serialization protocol RESP
2. Implemented a TCP connection that waits for two requests and responds with the message "PONG"

### Stage 4

1. Handled 2 concurrent requests and responded with resp(+PONG/r/n)
2. Implemented with the help of go-routines 🙃

### Stage 5

1. Handled concurrent requests and responded with custom message

### Stage 6

1. Implemented GET / SET in RDB

### Stage 7

1. Implemented GET / SET with px timeout command

> :rocket: Done with all the basic implementation of RDB
