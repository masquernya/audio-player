package server

import (
	"errors"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Shared struct {
	certMux sync.Mutex
}

func newShared() *Shared {
	s := &Shared{}
	return s
}

func (s *Shared) isPortInUse(p int) bool {
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(p))
	if err != nil {
		var s *net.OpError
		if errors.As(err, &s) {
			if s.Op == "dial" {
				return false
			}
			log.Fatal("error dialing:", err)
		} else {
			log.Fatal("error dialing:", err)
		}
	}
	conn.Close()
	return true
}

func (s *Shared) getFreePort() int {
	for port := DefaultPort; port < DefaultPort+100; port++ {
		if !s.isPortInUse(port) {
			return port
		}
		log.Println("port", port, "is in use")
		time.Sleep(time.Millisecond * 100)
	}
	log.Fatal("could not find free port")
	return -1
}

func (s *Shared) GetPort() int {
	portFilePath := path.Join(path.Dir(os.Args[0]), "port")
	bits, err := os.ReadFile(portFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("port file does not exist, finding free port")
		} else {
			log.Fatal("error stating port file:", err)
		}
	} else {
		str := strings.ReplaceAll(string(bits), "\n", "")
		port, err := strconv.Atoi(str)
		if err != nil {
			log.Fatal("error converting port file to int:", err)
		}
		return port
	}

	port := s.getFreePort()
	err = os.WriteFile(portFilePath, []byte(strconv.Itoa(port)), 0600)
	if err != nil {
		log.Fatal("error writing port file:", err)
	}
	return port
}
