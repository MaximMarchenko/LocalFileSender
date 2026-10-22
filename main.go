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

	file, err := os.Open(dirToFile)
	if err != nil {
		return err
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

func receive(addr, dirToFile string) error {
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

	dirToFile, err = makeDirAbs(dirToFile)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(dirToFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
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
		log.Fatal("not enogh args\nUsage send|receive, ipaddr, name of file. File can be asbsolute or relative path")
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
