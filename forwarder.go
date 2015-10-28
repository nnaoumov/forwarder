package main

//Simple TCP data forwarder that will forward any received data to a list of
//destinations ignoring any errors or return values

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

var argListenAddress *string
var argRemoteAddreses *string
var argBgSend *bool
var argTimeout *int

//var addresses = [...]string{"localhost:8000", "localhost:8001"}
var addresses []string
var timeout time.Duration

func loadArgs() {
	argBgSend = flag.Bool("bg-send", true, "ack to sender before forwarding to others")
	argListenAddress = flag.String("listen-adr", ":8082", "listen address")
	argRemoteAddreses = flag.String("remote-adr-list", "", "comma delimited list of remote hosts - e.g. host1:8082,host2:8082")
	argTimeout = flag.Int("timeout", 3, "connect timeout in seconds")

	flag.Parse()
	if *argRemoteAddreses == "" {
		fmt.Println("Error: remote-adr-list not specified")
		flag.Usage()
		os.Exit(1)
	}
	addresses = strings.Split(*argRemoteAddreses, ",")
	timeout = time.Duration(time.Duration(*argTimeout) * time.Second)
}

func main() {
	loadArgs()

	fmt.Printf("Starting server on %s\n", *argListenAddress)

	address, _ := net.ResolveTCPAddr("tcp", *argListenAddress)
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		fmt.Printf("ERROR: ListenTCP - %s", err)
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Printf("ERROR: AcceptTCP - %s\n", err)
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
		if *argBgSend {
			go send(addresses[i], buf)
		} else {
			send(addresses[i], buf)
		}
	}
	fmt.Println("handleRequest done")
}

func send(address string, buf []byte) {
	fmt.Printf("Sending data to %s\n", address)
	client, err := net.DialTimeout("tcp", address, timeout) //TODO add timeout
	if err != nil {
		fmt.Printf("Error connecting to %s - %s\n", address, err)
		return
	}
	defer client.Close()

	_, err = client.Write(buf)
	if err != nil {
		fmt.Printf("failed to write stats - %s\n", err)
		return
	}
	fmt.Printf("send to %s done\n", address)
}
