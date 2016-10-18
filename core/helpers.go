package core

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/foomo/htpasswd"
	"math/rand"
	"net"
	"os"
	"os/user"
	"strconv"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GetUser(username string) (*user.User, error) {
	if username == "" {
		curUser, err := user.Current()
		if err != nil {
			return nil, err
		}
		return curUser, nil
	} else {
		curUser, err := user.Lookup(username)
		if err != nil {
			return nil, err
		}
		return curUser, nil
	}
}

func GetRandom(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func CreateHttpAuth(file string, name string, password string) error {
	err := htpasswd.SetPassword(file, name, password, htpasswd.HashBCrypt)
	if err != nil {
		return err
	}
	err = os.Chmod(file, 0755)
	return err
}

func GetFreePort() (string, error) {
	rand.Seed(time.Now().UnixNano())

	log.Debugln("Attempting to get a free port")

	tries := 0
	for tries < 10 {
		port := strconv.Itoa(rand.Intn(50000) + 1024)
		log.Debugf("Claimed port %d", port)

		if PortFree(port) {
			return port, nil
		}

		tries++
	}
	return "", errors.New("Could not find free port.")
}

func PortFree(port string) bool {
	log.Debugf("Checking if port %s is free", port)

	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		log.Debugf("Port is free: %s", err.Error())
		return true
	} else {
		log.Debugf("Port already in use")
		conn.Close()
		return false
	}
}
