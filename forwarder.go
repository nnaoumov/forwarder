package main

//Simple TCP data forwarder that will forward any received data to a list of
//destinations ignoring any errors or return values

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"
)

const (
	TCP_READ_SIZE = 4096 * 1024
)

var tcpListenAddr = ":8082"
var addresses = [...]string{"localhost:8000", "localhost:8001"}
var timeout = time.Duration(3 * time.Second)

func main() {
	fmt.Printf("Starting server on %s", tcpListenAddr)

	address, _ := net.ResolveTCPAddr("tcp", tcpListenAddr)
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		fmt.Errorf("ERROR: ListenTCP - %s", err)
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Errorf("ERROR: AcceptTCP - %s", err)
			continue
		}

		fmt.Println("Accepted connection from " + conn.RemoteAddr().String())
		handleRequest(conn)
		fmt.Println("Finished processing connection from " + conn.RemoteAddr().String())
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	reqLen := len(buf)
	fmt.Printf("Received %d bytes\n", reqLen)

	for i := range addresses {
		go send(addresses[i], buf)
	}
	fmt.Println("handleRequest done")
}

func send(address string, buf []byte) {
	fmt.Printf("Sending data to %s\n", address)
	client, err := net.DialTimeout("tcp", address, timeout) //TODO add timeout
	if err != nil {
		fmt.Errorf("Error connecting to %s - %s", address, err)
		return
	}
	defer client.Close()

	_, err = client.Write(buf)
	if err != nil {
		fmt.Errorf("failed to write stats - %s", err)
		return
	}
	fmt.Printf("send to %s done\n", address)
}
