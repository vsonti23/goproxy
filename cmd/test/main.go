package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}

	conn, err := listener.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	data, err := io.ReadAll(conn)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Got EOF after receiving: %q\n", data)

	// We received FIN, but our write direction is still open.
	fmt.Println("Waiting 5 seconds...")
	time.Sleep(5 * time.Second)

	fmt.Println("Sending response after EOF")
	conn.Write([]byte("backend response\n"))
}
