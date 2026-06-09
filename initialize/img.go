package initialize

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"

	"github.com/oldfritter/lucy/lib/captcha"
)

func init() {
	directory := "config/img"
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
		if fileInfo.IsDir() {
			continue
		}
		path := fmt.Sprintf("%s/%s", directory, fileInfo.Name())
		f, err := os.Open(path)
		if err != nil {
			log.Printf("跳过图片 %s: %v", fileInfo.Name(), err)
			continue
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			log.Printf("跳过图片 %s: %v", fileInfo.Name(), err)
			continue
		}
		captcha.BackgroundImgs = append(captcha.BackgroundImgs, img)
	}
	if len(captcha.BackgroundImgs) == 0 {
		fmt.Println("warning: no background images loaded for rotate captcha")
	}
}
