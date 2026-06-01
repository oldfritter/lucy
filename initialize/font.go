package initialize

import (
	"fmt"
	"log"
	"os"

	"github.com/golang/freetype"

	"github.com/oldfritter/lucy/lib/captcha"
)

func init() {
	directory := "config/ttf"
	files, err := os.Open(directory)
	if err != nil {
		fmt.Println("error opening directory:", err)
		return
	}
	defer files.Close()
	fileInfos, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err)
		return
	}
	for _, fileInfo := range fileInfos {
		fontBytes, err := os.ReadFile(fmt.Sprintf("%s/%s", directory, fileInfo.Name()))
		if err != nil {
			log.Fatalf("Error reading font file: %v", err)
		}
		f, err := freetype.ParseFont(fontBytes)
		if err != nil {
			log.Printf("跳过字体 %s: %v", fileInfo.Name(), err)
			continue
		}
		captcha.Fonts = append(captcha.Fonts, f)
	}
}
