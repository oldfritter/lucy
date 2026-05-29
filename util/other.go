package util

import (
	"crypto/rand"
	"log"
	"math/big"
	"os"
	"strings"
)

func RandStringRunes(n int) string {
	var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890")
	b := make([]rune, n)
	for i := range b {
		index, _ := rand.Int(rand.Reader, big.NewInt(62))
		b[i] = letterRunes[index.Int64()]
	}
	return string(b)
}

func RandNumberStringRunes(n int) string {
	var letterRunes = []rune("1234567890")
	b := make([]rune, n)
	for i := range b {
		index, _ := rand.Int(rand.Reader, big.NewInt(10))
		b[i] = letterRunes[index.Int64()]
	}
	return string(b)
}

func GetLogFile(name ...string) *os.File {
	var app string
	if len(name) == 0 {
		exe := strings.Split(os.Args[0], "/")
		app = exe[len(exe)-1]
	} else {
		app = name[0]
	}
	if err := os.Mkdir("log", 0755); err != nil {
		if !os.IsExist(err) {
			log.Fatalf("create folder error: %v", err)
		}
	}
	file, err := os.OpenFile("log/"+app+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
	return file
}

func Title(in string) string {
	var out []byte
	for i, c := range []byte(in) {
		if i > 0 {
			out = append(out, c)
			continue
		}
		if 64 < c && c < 91 {
			out = append(out, c)
		}
		if 96 < c && c < 123 {
			out = append(out, c-32)
		}
	}
	return string(out)
}
