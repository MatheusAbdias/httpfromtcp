package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer f.Close()
		defer close(ch)
		lineContent := ""
		buffer := make([]byte, 8)
		for {
			size, err := f.Read(buffer)
			content := string(buffer[:size])
			if err == io.EOF {
				return
			} else if err != nil {
				log.Fatal(err)
			}

			parts := strings.Split(content, "\n")
			lineContent += parts[0]
			if len(parts) > 1 {
				ch <- lineContent
				lineContent = parts[1]
			}
		}
	}()

	return ch
}

func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Println("connection has accepted")

		ch := getLinesChannel(conn)
		for line := range ch {
			fmt.Println(line)
		}

		fmt.Println("connection closed")
	}
}
