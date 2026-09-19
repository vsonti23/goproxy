package main

import (
	"fmt"
	"io"
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

	buffer := make([]byte, 4)

	for {
		n, err := conn.Read(buffer)

		if err == io.EOF {
			fmt.Println("Client Disconnected")
			break
		}

		if err != nil {
			panic(err)
		}

		fmt.Printf("Read %d bytes: %q\n", n, string(buffer[:n]))
	}	
}
