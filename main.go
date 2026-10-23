package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func makeDirAbs(dir string) (string, error) {
	if err := error(nil); strings.Index(dir, ".") == 0 || strings.Index(dir, string(os.PathSeparator)) != 0 {
		dir, err = filepath.Abs(dir)
		if err != nil {
			return "", err
		}
	}
	return dir, nil
}

func send(addr, dirToFile string) error {
	dirToFile, err := makeDirAbs(dirToFile)
	if err != nil {
		return err
	}

	fmt.Print("waiting for connection...")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Println("connected")
	file, err := os.Open(dirToFile)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Print("starting to send file")

	if _, err := io.Copy(conn, file); err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("sent file successfully")

	return nil
}

func receive(addr, dirToFile string) error {
	dirToFile, err := makeDirAbs(dirToFile)
	if err != nil {
		return err
	}

	fmt.Print("waiting for connection...")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	fmt.Println("connected")

	file, err := os.OpenFile(dirToFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Print("writing data into file...")

	if _, err := io.Copy(file, conn); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("received file successfully")

	return nil
}

func main() {
	if len(os.Args) < 4 {
		log.Fatal("not enough args\nUsage send|receive, ipaddr, name of file. File can be absolute or relative path")
	}
	switch os.Args[1] {
	case "send":
		if err := send(os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}
	case "receive":
		if err := receive(os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("unknown command \nUsage send|receive, ipaddr, name of file. File can be absolute or relative path")
	}
}
