package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	
	fmt.Println(("Listening on :8080"))

	conn, err := listener.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Client connected:", conn.RemoteAddr())
}
