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
	
	fmt.Println("Listening on :8080")

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

	done := make(chan struct{}, 2)

	go func() {
		copyData(backend, conn)
		done <- struct{}{}
	}()

	go func() {
		copyData(conn, backend)
		done <- struct{}{}
	}()

	<-done
}

func copyData(dst net.Conn, src net.Conn) {
	buffer := make([]byte, 1024)

	for {
		n, err := src.Read(buffer)
		if err == io.EOF {
			fmt.Println("Connection Closed")
			break
		}
		
		if err != nil {
			fmt.Println("Failed to read:", err)
			break
		}

		fmt.Printf("Read: %q\n", string(buffer[:n]))

		_, err = dst.Write(buffer[:n])
		if err != nil {
			fmt.Println("Failed to write:", err)
			break
		}
	}
}
