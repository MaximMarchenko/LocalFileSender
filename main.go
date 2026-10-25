package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FileInfo struct {
	Name string `json:"name"`
	Size uint64 `json:"size"`
}

func makeCopyFile(name string, number int) string {
	baseName := name[0:strings.LastIndex(name, ".")]
	nameExt := filepath.Ext(name)
	return fmt.Sprintf("%s_COPY_%d%s", baseName, number, nameExt)
}

func getInfo(file string) (*FileInfo, error) {
	info, err := os.Stat(file)
	if err != nil {
		return &FileInfo{}, err
	}
	return &FileInfo{Name: info.Name(), Size: uint64(info.Size())}, nil
}

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

	fmt.Println()
	resp := make([]byte, 2)
	_, err = conn.Read(resp)
	if err != nil {
		return err
	}
	if strings.ToLower(string(resp)) != "ok" {
		fmt.Println()
		return fmt.Errorf("access denied")
	}

	fmt.Println("connected")
	info, err := getInfo(dirToFile)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(info)
	if err != nil {
		return err
	}
	_, err = conn.Write(jsonData)
	if err != nil {
		return err
	}
	for {
		_, err = conn.Read(resp)
		if err != nil {
			return err
		}
		if strings.ToLower(string(resp)) == "ok" {
			break
		} else {
			return fmt.Errorf("another device denied")
		}
	}
	file, err := os.OpenFile(dirToFile, os.O_RDONLY, 0644)
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
	fmt.Println()
	fmt.Printf("some device wants to connect, IP: %s, accept ? [y/N]: ", conn.RemoteAddr())
	answer := ""
	fmt.Scan(&answer)
	switch strings.ToLower(answer) {
	case "y":
		conn.Write([]byte("OK"))
	default:
		conn.Write([]byte("NO"))
		return fmt.Errorf("access denied for device IP: %s", conn.RemoteAddr())
	}

	fmt.Println("connected")
	jsonData := make([]byte, 30)
	info := FileInfo{}
	for {
		_, err = conn.Read(jsonData)
		if err != nil {
			conn.Write([]byte("NO"))
			return err
		}
		err = json.Unmarshal(jsonData, &info)
		if err != nil {
			conn.Write([]byte("NO"))
			return err
		}
		if info == (FileInfo{Name: "", Size: uint64(0)}) {
			conn.Write([]byte("NO"))
			return fmt.Errorf("empty data")
		} else {
			conn.Write([]byte("OK"))
			break
		}
	}

	fileName := filepath.Base(dirToFile)
	if info.Name != fileName {
		fmt.Printf("choose which file to write, %q %q ? [1/2]", fileName, info.Name)
		fmt.Scan(&answer)
		num, _ := strconv.Atoi(answer)
		switch num {
		case 1:
		case 2:
			dirToFile, err = makeDirAbs(info.Name)
			if err != nil {
				return err
			}
		default:
			fmt.Println()
			fmt.Println("answer out of range")
		}
	}

	dir, err := makeDirAbs(".")
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	copies := make([]string, 0)
	nameFile := info.Name[0:strings.LastIndex(info.Name, ".")]
	for _, en := range entries {
		if en.IsDir() {
			continue
		}
		if strings.Contains(en.Name(), nameFile) && en.Name()[0] == info.Name[0] && filepath.Ext(nameFile) == filepath.Ext(en.Name()) {
			copies = append(copies, en.Name())
		}
	}
	if len(copies) > 0 {
		dirToFile, err = makeDirAbs(makeCopyFile(info.Name, len(copies)))
		if err != nil {
			return err
		}
	}

	flagsToWrite := os.O_CREATE | os.O_RDWR
	if _, err = os.Stat(dirToFile); err == nil {
		fmt.Printf("append data to file %q ? [Y/n]: ", dirToFile)
		fmt.Scan(&answer)
		switch strings.ToLower(answer) {
		case "n":
		default:
			flagsToWrite |= os.O_APPEND
		}
	}

	file, err := os.OpenFile(dirToFile, flagsToWrite, 0644)
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
