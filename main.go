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

	done := make(chan error, 2)

	go func() {
		done <- copyData(backend, conn)
	}()

	go func() {
		done <- copyData(conn, backend)
	}()

	connErr := <-done

	fmt.Println("Closing Connection")

	if connErr != nil {
		fmt.Println("Proxy connection ended with error:", connErr)
	}
}

func copyData(dst net.Conn, src net.Conn) error {
	buffer := make([]byte, 1024)

	for {
		n, err := src.Read(buffer)
		if err == io.EOF {
			return nil
		}
		
		if err != nil {
			return err
		}

		fmt.Printf("Read: %q\n", string(buffer[:n]))

		_, err = dst.Write(buffer[:n])
		if err != nil {
			return err
		}
	}
}
