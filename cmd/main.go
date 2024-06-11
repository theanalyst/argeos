package main

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection: ", err)
	}

	message := string(buf[:n])
	fmt.Println("Received message ", message)
}

func startServer(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error starting listener: ", err)
		return
	}

	defer listener.Close()
	fmt.Println("Starting argeos daemon on ", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}
		go handleConnection(conn)
	}

}


func main() {
	startServer(":9999")
}
