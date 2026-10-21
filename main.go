package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func send(addr, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer file.Close()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := io.Copy(conn, file); err != nil {
		return err
	}

	return nil
}

func receive(addr, fileName string) error {
	listenr, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listenr.Close()

	conn, err := listenr.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, conn); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) < 4 {
		log.Fatal("not enogh args\nUsage send|receive, ipaddr, name of file. File must be near running program|will create near running program")
	}
	if strings.ToLower(os.Args[1]) == "send" {
		fmt.Println("starting to connect and sening")
		if err := send(os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}
	} else if strings.ToLower(os.Args[1]) == "receive" {
		if err := receive(os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("unknown command")
	}

}
