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

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}	
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Client connected:", conn.RemoteAddr())

	backend, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println("Failed to connect to backend", err)
		return
	}
	defer backend.Close()

	fmt.Println("Connected to backend:", backend.RemoteAddr())

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

		fmt.Printf("Read: %q\n", string(buffer[:n]))

		_, err = backend.Write(buffer[:n])
		if err != nil {
			fmt.Println("Failed to write to backend:", err)
			break
		}
	}	
}
