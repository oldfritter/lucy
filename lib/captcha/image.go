package captcha

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"

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
	centerX := float64(bounds.Dx() / 2)
	centerY := float64(bounds.Dy() / 2)
	for i := 0; i < bounds.Dx(); i++ {
		for j := 0; j < bounds.Dy(); j++ {
			dx := float64(i) - centerX
			dy := float64(j) - centerY
			radius := math.Sqrt(dx*dx + dy*dy)
			theta := math.Atan2(dy, dx)
			newRadius := radius * scale
			newX := int(centerX + newRadius*math.Cos(theta+angle))
			newY := int(centerY + newRadius*math.Sin(theta+angle))
			if newX >= 0 && newX < bounds.Dx() && newY >= 0 && newY < bounds.Dy() {
				dst.Set(i, j, img.At(newX, newY))
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
		Dot:  fixed.P(x, y),
	}
	drawer.DrawString(text)
	return newImg
}

// GenerateCaptchaImage 生成验证码图片（透明背景，动态尺寸）
func GenerateCaptchaImage(prompts []string) image.Image {
	var texts []image.Image
	for _, p := range prompts {
		texts = append(texts, CreateTextImage(p))
	}

	n := len(prompts)
	w := 40 + n*80
	h := 120

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	AddText(img, "请依次点击以下文字：", 10, 30)
	for i, t := range texts {
		AddImage(img, t, 100+30*i, 30, 10)
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
		Dot:  fixed.P(0, 25),
	}
	drawer.DrawString(text)
	return img
}

// GenerateTextChallenge 将一组字符分列网格放置（含随机旋转 ±60°），返回图片及各字符的中心坐标
func GenerateTextChallenge(chars []string) (image.Image, []image.Point) {
	fontSize := 60.0
	fontSizeInt := int(fontSize)
	padding := 20

	// 布局间距独立于渲染字号，保持画布紧凑
	layoutSize := 20
	expansion := layoutSize*2 + 10

	n := len(chars)
	w := padding*2 + n*expansion*3
	h := padding*2 + expansion*4

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	xMax := w - padding - expansion
	yMin := padding + expansion
	yMax := h - padding - expansion

	var points []image.Point

	// 单字临时画布尺寸（按实际渲染字号，足够容纳旋转后的字形）
	charSz := fontSizeInt * 3

	// 网格布局：将画布等分为 n 列，每列中心放一个字，可在列内随机偏移
	colW := (xMax - padding) / n
	colH := yMax - yMin
	colJitterX := max(colW/2-charSz/2, 0)
	colJitterY := max(colH/2-charSz/2, 0)

	for i, ch := range chars {
		colCenterX := padding + colW*i + colW/2
		colCenterY := yMin + colH/2

		x := colCenterX + rand.Intn(colJitterX*2+1) - colJitterX
		y := colCenterY + rand.Intn(colJitterY*2+1) - colJitterY

		angleDeg := rand.Float64()*120 - 60 // [-60, 60]

		points = append(points, image.Point{X: x, Y: y})

		// 创建单字图片（透明背景），居中绘制
		charImg := image.NewRGBA(image.Rect(0, 0, charSz, charSz))
		fg := image.NewUniform(color.RGBA{
			uint8(rand.Intn(180)),
			uint8(rand.Intn(180)),
			uint8(rand.Intn(180)),
			255,
		})
		drawer := font.Drawer{
			Dst:  charImg,
			Src:  fg,
			Face: truetype.NewFace(getFont(), &truetype.Options{Size: fontSize}),
			Dot:  fixed.P(charSz/2-fontSizeInt/3, charSz/2+fontSizeInt/3),
		}
		drawer.DrawString(ch)

		// 旋转单字图片
		rotated := ImageRotate(charImg, angleDeg)
		rb := rotated.Bounds()

		// 将旋转后的字形合成到主画布（中心对齐到 x, y）
		compositePoint := image.Pt(x-rb.Dx()/2, y-rb.Dy()/2)
		draw.Draw(img,
			image.Rect(compositePoint.X, compositePoint.Y, compositePoint.X+rb.Dx(), compositePoint.Y+rb.Dy()),
			rotated,
			rb.Min,
			draw.Over,
		)
	}

	return img, points
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
