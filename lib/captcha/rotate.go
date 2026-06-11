package captcha

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// ImageRotate 将图片按指定角度旋转（逆时针为正），双线性插值，返回旋转后图像
func ImageRotate(src image.Image, angle float64) image.Image {
	bounds := src.Bounds()
	rad := angle * math.Pi / 180.0

	w, h := float64(bounds.Dx()), float64(bounds.Dy())
	cos, sin := math.Abs(math.Cos(rad)), math.Abs(math.Sin(rad))
	nw := int(w*cos + h*sin)
	nh := int(w*sin + h*cos)

	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))

	cx, cy := w/2, h/2
	ncx, ncy := float64(nw)/2, float64(nh)/2

	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			nx := float64(x) - ncx
			ny := float64(y) - ncy

			sx := nx*math.Cos(rad) + ny*math.Sin(rad) + cx
			sy := -nx*math.Sin(rad) + ny*math.Cos(rad) + cy

			if sx < 0 || sx >= w-1 || sy < 0 || sy >= h-1 {
				continue
			}

			dst.Set(x, y, bilinearSample(src, sx, sy))
		}
	}

	return dst
}

// bilinearSample 对源图 (sx, sy) 做双线性插值采样
func bilinearSample(src image.Image, sx, sy float64) color.RGBA {
	x0 := int(math.Floor(sx))
	y0 := int(math.Floor(sy))
	x1 := x0 + 1
	y1 := y0 + 1

	fx := sx - float64(x0)
	fy := sy - float64(y0)

	c00 := rgba(src.At(x0, y0))
	c10 := rgba(src.At(x1, y0))
	c01 := rgba(src.At(x0, y1))
	c11 := rgba(src.At(x1, y1))

	// 双线性插值
	R := uint8((1-fy)*((1-fx)*float64(c00.R)+fx*float64(c10.R)) +
		fy*((1-fx)*float64(c01.R)+fx*float64(c11.R)))
	G := uint8((1-fy)*((1-fx)*float64(c00.G)+fx*float64(c10.G)) +
		fy*((1-fx)*float64(c01.G)+fx*float64(c11.G)))
	B := uint8((1-fy)*((1-fx)*float64(c00.B)+fx*float64(c10.B)) +
		fy*((1-fx)*float64(c01.B)+fx*float64(c11.B)))
	A := uint8((1-fy)*((1-fx)*float64(c00.A)+fx*float64(c10.A)) +
		fy*((1-fx)*float64(c01.A)+fx*float64(c11.A)))

	return color.RGBA{R: R, G: G, B: B, A: A}
}

// rgba 将 color.Color 转为 color.RGBA
func rgba(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

// GenerateRotateCaptcha 生成旋转验证码图片。
// 从系统图片库随机选取一张背景图，裁剪为直径 200px 的圆形，随机旋转，
// 最后以图片圆心为中心裁剪为 200×200。
// 返回：旋转并裁剪后的图片、逆时针旋转角度
func GenerateRotateCaptcha() (image.Image, int) {
	bg := getBackground()

	// 将背景图裁剪为直径 200px 的圆形
	circle := toCircle(bg)

	angle := int(math.Round(120 + rand.Float64()*180))

	rotated := ImageRotate(circle, float64(angle))
	cropped := centerCrop(rotated, 200)
	return cropped, angle
}

// centerCrop 从图片中心裁取指定尺寸的区域
func centerCrop(src image.Image, size int) *image.RGBA {
	bounds := src.Bounds()
	cx := (bounds.Dx() - size) / 2
	cy := (bounds.Dy() - size) / 2
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dst.Set(x, y, src.At(cx+x, cy+y))
		}
	}
	return dst
}

// toCircle 将图片缩放并裁剪为直径 200px 的圆形（圆形外透明）
func toCircle(src image.Image) *image.RGBA {
	const diameter = 200
	radius := float64(diameter) / 2
	dst := image.NewRGBA(image.Rect(0, 0, diameter, diameter))

	srcBounds := src.Bounds()
	srcW := float64(srcBounds.Dx())
	srcH := float64(srcBounds.Dy())

	// 缩放比例：保持宽高比覆盖 200x200（超出部分中心裁切）
	scale := math.Max(float64(diameter)/srcW, float64(diameter)/srcH)
	sw := srcW * scale
	sh := srcH * scale
	ox := (sw - float64(diameter)) / 2
	oy := (sh - float64(diameter)) / 2

	for y := 0; y < diameter; y++ {
		for x := 0; x < diameter; x++ {
			// 圆形遮罩：圆外像素保持透明
			dx := float64(x) - radius
			dy := float64(y) - radius
			if dx*dx+dy*dy > radius*radius {
				continue
			}

			// 映射回原图坐标
			sx := (float64(x) + ox) / scale
			sy := (float64(y) + oy) / scale

			if sx < 0 || sx >= srcW || sy < 0 || sy >= srcH {
				continue
			}

			dst.Set(x, y, bilinearSample(src, sx, sy))
		}
	}

	return dst
}

// getBackground 从系统图片库随机选取一张背景图
func getBackground() image.Image {
	if len(BackgroundImgs) == 0 {
		// 无背景图时回退到白色画布
		img := image.NewRGBA(image.Rect(0, 0, 300, 300))
		for x := 0; x < 300; x++ {
			for y := 0; y < 300; y++ {
				img.Set(x, y, color.White)
			}
		}
		return img
	}
	return BackgroundImgs[rand.Intn(len(BackgroundImgs))]
}
