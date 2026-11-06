package main

import (
	"log"
	"os"

	"github.com/MaximMarchenko/LocalFileSender/internal/transfer"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatal("not enough args\nUsage send|receive, ipaddr, name of file. File can be absolute or relative path")
	}
	switch os.Args[1] {
	case "send":
		if err := transfer.Send(os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}
	case "receive":
		if err := transfer.Receive(os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("unknown command \nUsage send|receive, ipaddr, name of file. File can be absolute or relative path")
	}
}
