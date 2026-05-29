package captcha

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	Fonts []*truetype.Font
)

func getFont() *truetype.Font {
	return Fonts[int(rand.Intn(len(Fonts)))]
}

// 旋转图像
func Whirl(img image.Image, angle, scale float64) image.Image {
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)
	// 计算旋转中心点
	centerX := float64(bounds.Dx() / 2)
	centerY := float64(bounds.Dy() / 2)
	for i := 0; i < bounds.Dx(); i++ {
		for j := 0; j < bounds.Dy(); j++ {
			// 计算当前像素点到中心点的距离
			dx := float64(i) - centerX
			dy := float64(j) - centerY
			// 根据旋转角度计算变形后的像素点坐标
			radius := math.Sqrt(dx*dx + dy*dy)
			theta := math.Atan2(dy, dx)
			newRadius := radius * scale
			newX := int(centerX + newRadius*math.Cos(theta+angle))
			newY := int(centerY + newRadius*math.Sin(theta+angle))
			// 如果新的像素点坐标在图像范围内，则进行变形处理
			if newX >= 0 && newX < bounds.Dx() && newY >= 0 && newY < bounds.Dy() {
				// 获取原图像素点颜色
				color := img.At(newX, newY)
				// 绘制新图像
				dst.Set(i, j, color)
			}
		}
	}
	return dst
}

// 图片上添加文字
func AddText(img image.Image, text string, x, y int) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	draw.Draw(newImg, bounds, img, bounds.Min, draw.Src)
	fg := image.NewUniform(color.RGBA{
		uint8(rand.Int63n(255)),
		uint8(rand.Int63n(255)),
		uint8(rand.Int63n(255)),
		255,
	})
	drawer := font.Drawer{
		Dst:  newImg,
		Src:  fg,
		Face: truetype.NewFace(getFont(), &truetype.Options{Size: 30}),
		Dot:  fixed.P(x, y), // 文字起始位置
	}
	drawer.DrawString(text)
	return newImg
}

// LoadBackground 加载背景图片
func LoadBackground(inputPath string) image.Image {
	reader, err := os.Open(inputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	m, _, err := image.Decode(reader)
	return m
}

// GenerateCaptchaImage 生成验证码图片：背景 + 提示文字 + 逐个叠加文字图片
func GenerateCaptchaImage(prompts []string, backgroundPath string) image.Image {
	var texts []image.Image
	for _, p := range prompts {
		texts = append(texts, CreateTextImage(p))
	}

	img := LoadBackground(backgroundPath)
	AddText(img, "请依次点击以下文字：", 0, 30)
	for i, t := range texts {
		AddImage(img, t, 100+30*i, 30, 0)
	}
	return img
}

// 生成文字图片
func CreateTextImage(text string) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 30, 30))
	draw.Draw(img, img.Bounds(), image.Transparent, image.Point{}, draw.Src)
	fg := image.NewUniform(color.RGBA{
		uint8(rand.Int63n(255)),
		uint8(rand.Int63n(255)),
		uint8(rand.Int63n(255)),
		255,
	})
	drawer := font.Drawer{
		Dst:  img,
		Src:  fg,
		Face: truetype.NewFace(getFont(), &truetype.Options{Size: 30}),
		Dot:  fixed.P(0, 25), // 文字起始位置
	}
	drawer.DrawString(text)
	return img
}

// 叠加图片
func AddImage(img, addImg image.Image, x, y, offset int) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	draw.Draw(newImg, bounds, img, bounds.Min, draw.Src)
	draw.Draw(
		newImg,
		bounds,
		addImg,
		image.Pt(-x, -y-offset),
		draw.Over,
	)
	return newImg
}
