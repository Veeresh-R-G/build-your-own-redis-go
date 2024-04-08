## Basic ReDis Setup

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

## Replication Setup for Redis

### Stage - 1
1. Run the redis server on custom port specified using the --port flag

### Stage - 2
1. Added support for INFO REPLICATION COMMAND

### Stage - 3 / 4
1. Added support to return the Master Node ID and Offset

### Stage - 5 Send Handshake - 1 / 3
1. Pinged the Master node :)
2. I'm yet to Refactor the code

### Stage - 6 Send Handshake - 2 / 3
1. Received and pinged master twice again :)

### Stage - 7 Send Handshake - 3 / 3
1. Relayed the PSYNC to master node
> :rocket: SEND Handshake done

### Stage - 8 Receive Handshake - 1 / 2
1. Relayed back message acting as a master node. (for REPLCONF)

### Stage - 9 Receive Handshake - 2 / 2
1. Relayed back message for PSYNC. Relayed FULLSYNC which tells the worker node that master is unable to perform incremental synchronization so it will perform full synchronization with the worker.
> :rocket: Receive Handshake done

