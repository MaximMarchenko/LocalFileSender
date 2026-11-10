package transfer

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MaximMarchenko/LocalFileSender/internal/json"
	"github.com/MaximMarchenko/LocalFileSender/internal/manifest"
)

const (
	msgOk       = "OK"
	msgFail     = "FAIL"
	msgFinish   = "FINISH"
	msgNext     = "NEXT"
	msgAgain    = "AGAIN"
	msgCheckSum = "CHECKSUM"
)

func readConn(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\n"), nil
}

func writeConn(w *bufio.Writer, line string) error {
	if _, err := w.WriteString(line + "\n"); err != nil {
		return err
	}
	return w.Flush()
}

func sendFiles(path string, rw *bufio.ReadWriter, conn net.Conn, current *int, total int, m *manifest.Manifest) error {
	if len(m.Entries) > 0 {
		for _, en := range m.Entries {
			if err := sendFiles(filepath.Join(path, m.Name), rw, conn, current, total, en); err != nil {
				return err
			}
		}
	} else {
		*current++
		isLast := *current == total
		hash := sha256.New()
		file, err := os.OpenFile(filepath.Join(path, m.Name), os.O_RDONLY, 0o644)
		if err != nil {
			return err
		}
		writeConn(rw.Writer, filepath.Join(path, m.Name))
		if resp, _ := readConn(rw.Reader); resp != msgOk {
			return fmt.Errorf("issue to send files")
		}
		writeConn(rw.Writer, strconv.Itoa(int(m.Size)))
		if resp, _ := readConn(rw.Reader); resp != msgOk {
			return fmt.Errorf("issue to send files")
		}
		for {
			hash.Reset()
			file.Seek(0, io.SeekStart)
			if _, err = io.Copy(io.MultiWriter(conn, hash), file); err != nil {
				file.Close()
				return err
			}
			for {
				writeConn(rw.Writer, msgCheckSum)
				if resp, _ := readConn(rw.Reader); resp != msgOk {
					time.Sleep(10 * time.Millisecond)
				} else {
					break
				}
			}
			writeConn(rw.Writer, hex.EncodeToString(hash.Sum(nil)))
			if resp, _ := readConn(rw.Reader); resp != msgAgain {
				if isLast {
					writeConn(rw.Writer, msgFinish)
				} else {
					writeConn(rw.Writer, msgNext)
				}
				if resp, _ = readConn(rw.Reader); resp != msgOk {
					return fmt.Errorf("issue to send files")
				}
				break
			}
		}
	}
	return nil
}

func receiveFiles(path string, rw *bufio.ReadWriter, conn net.Conn) error {
	hash := sha256.New()
receiving:
	for {
		dir, err := readConn(rw.Reader)
		if err != nil {
			writeConn(rw.Writer, msgFail)
			return err
		}
		writeConn(rw.Writer, msgOk)
		size, err := readConn(rw.Reader)
		S, err := strconv.Atoi(size)
		if err != nil {
			writeConn(rw.Writer, msgFail)
			return err
		}
		writeConn(rw.Writer, msgOk)
		file, err := os.OpenFile(filepath.Join(path, dir), os.O_CREATE|os.O_RDWR, 0o644)
		hash.Reset()
		file.Truncate(0)
		file.Seek(0, io.SeekStart)
		if _, err := io.CopyN(io.MultiWriter(hash, file), conn, int64(S)); err != nil {
			return err
		}
		if req, _ := readConn(rw.Reader); req != msgCheckSum {
			return fmt.Errorf("unknown request")
		} else {
			writeConn(rw.Writer, msgOk)
		}
		senderHash, err := readConn(rw.Reader)
		if err != nil {
			return err
		}
		if senderHash == hex.EncodeToString(hash.Sum(nil)) {
			writeConn(rw.Writer, msgOk)
			if err = file.Sync(); err != nil {
				return err
			}
			req, _ := readConn(rw.Reader)
			switch req {
			case msgNext:
				writeConn(rw.Writer, msgOk)
				continue
			case msgFinish:
				writeConn(rw.Writer, msgOk)
				break receiving
			default:
				writeConn(rw.Writer, msgFail)
				return fmt.Errorf("unknown request")
			}
		} else {
			writeConn(rw.Writer, msgAgain)
			continue
		}
	}
	return nil
}

func Send(addr, dirToFile string) error {
	dirToFile, err := manifest.ToAbsDir(dirToFile)
	if err != nil {
		return err
	}

	fmt.Print("waiting for connection...")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	fmt.Println()
	if resp, err := readConn(rw.Reader); resp != msgOk || err != nil {
		fmt.Println()
		return fmt.Errorf("access denied")
	}

	fmt.Println("connected")
	m, err := manifest.Build(dirToFile)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = writeConn(rw.Writer, string(jsonData))
	if err != nil {
		return err
	}
	if resp, err := readConn(rw.Reader); resp != msgOk || err != nil {
		return fmt.Errorf("another device denied")
	}

	fmt.Print("starting to send file")

	if m.HasFiles() || !m.IsDir {
		current := 0
		total := m.CountFiles()
		if err = sendFiles("", rw, conn, &current, total, m); err != nil {
			return err
		}
	}

	fmt.Println()
	fmt.Println("sent file successfully")

	return nil
}

func Receive(addr, dirToFile string) error {
	dirToFile, err := manifest.ToAbsDir(dirToFile)
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
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	switch strings.ToLower(answer) {
	case "y":
		writeConn(rw.Writer, msgOk)
	default:
		writeConn(rw.Writer, msgFail)
		return fmt.Errorf("access denied for device IP: %s", conn.RemoteAddr())
	}

	fmt.Println("connected")
	jsonData, err := readConn(rw.Reader)
	if err != nil {
		writeConn(rw.Writer, msgFail)
		return err
	}
	m, err := json.Unmarshal([]byte(jsonData))
	if err != nil {
		writeConn(rw.Writer, msgFail)
		return err
	}
	if m.Name == "" {
		return fmt.Errorf("empty data")
	} else {
		writeConn(rw.Writer, msgOk)
	}
	fileName := filepath.Base(dirToFile)
	if m.Name != fileName {
		fmt.Printf("choose which file to write, %q or %q ? [1/2]:", fileName, m.Name)
		fmt.Scan(&answer)
		num, _ := strconv.Atoi(answer)
		switch num {
		case 1:
		case 2:
			dirToFile, err = manifest.ToAbsDir(m.Name)
			if err != nil {
				return err
			}
		default:
			fmt.Println()
			fmt.Println("answer out of range")
		}
	}

	fmt.Print("writing data into file...")
	dirs, err := m.GetDirs(dirToFile)
	if len(dirs) > 0 {
		for _, d := range dirs {
			err = os.MkdirAll(d, 0o755)
		}
	}

	if m.HasFiles() || !m.IsDir {
		if err = receiveFiles(dirToFile, rw, conn); err != nil {
			return err
		}
	}

	fmt.Println()
	fmt.Println("received file successfully")

	return nil
}
