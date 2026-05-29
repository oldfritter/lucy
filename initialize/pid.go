package initialize

import (
	"log"
	"os"
	"strconv"
	"strings"
)

func init() {
	if err := os.Mkdir("pid", 0755); err != nil {
		if !os.IsExist(err) {
			log.Fatalf("create folder error: %v", err)
		}
	}
	exe := strings.Split(os.Args[0], "/")
	app := exe[len(exe)-1]
	err := os.WriteFile("pid/"+app+".pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
