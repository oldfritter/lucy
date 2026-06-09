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
	Fonts           []*truetype.Font
	BackgroundImgs  []image.Image // 旋转验证码背景图库
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

// charRect 表示一个已放置字符的边界矩形（含间距）
type charRect struct {
	MinX, MinY, MaxX, MaxY int
}

// overlaps 检查两个矩形是否重叠
func (r charRect) overlaps(other charRect) bool {
	return r.MinX < other.MaxX && r.MaxX > other.MinX &&
		r.MinY < other.MaxY && r.MaxY > other.MinY
}

// GenerateTextChallenge 将一组字符随机散布（含随机旋转 ±60°）在画布上，返回图片及各字符的中心坐标
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

	// 碰撞检测半径与字号匹配，防止实际渲染后的字形重叠
	collisionExpansion := fontSizeInt + 20

	var points []image.Point
	var placed []charRect
	const maxRetries = 50

	// 单字临时画布尺寸（按实际渲染字号，足够容纳旋转后的字形）
	charSz := fontSizeInt * 3

	// 分两行分布（n≥4），增加水平间距，减少垂直重叠
	rows := 1
	if n >= 4 {
		rows = 2
	}
	rowH := (yMax - yMin) / rows

	for i, ch := range chars {
		// 确定当前字所属的行 Y 范围
		rowIdx := i % rows
		yLo := yMin + rowIdx*rowH
		yHi := yMin + (rowIdx+1)*rowH
		if rowIdx == rows-1 {
			yHi = yMax
		}

		var x, y int
		var angleDeg float64
		for retry := 0; retry < maxRetries; retry++ {
			x = padding + rand.Intn(xMax-padding)
			y = yLo + rand.Intn(max(yHi-yLo, 1))
			angleDeg = rand.Float64()*120 - 60 // [-60, 60]

			candidate := charRect{
				MinX: x - collisionExpansion,
				MaxX: x + collisionExpansion,
				MinY: y - collisionExpansion,
				MaxY: y + collisionExpansion,
			}

			overlap := false
			for _, p := range placed {
				if candidate.overlaps(p) {
					overlap = true
					break
				}
			}
			if !overlap {
				break
			}
		}

		points = append(points, image.Point{X: x, Y: y})
		placed = append(placed, charRect{
			MinX: x - collisionExpansion,
			MaxX: x + collisionExpansion,
			MinY: y - collisionExpansion,
			MaxY: y + collisionExpansion,
		})

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
