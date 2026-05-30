package captcha

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// RotateImage 将图片按指定角度旋转（逆时针为正），双线性插值，返回旋转后图像
func RotateImage(src image.Image, angle float64) image.Image {
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

// GenerateRotateCaptcha 生成旋转验证码图片（透明背景，动态尺寸）。
// indicator: 方向指示文字，如 "▲"
// 返回：旋转后的图片、旋转角度（0-360）
func GenerateRotateCaptcha(indicator string) (image.Image, float64) {
	fontSize := 72.0
	w, h := 300, 300

	canvas := image.NewRGBA(image.Rect(0, 0, w, h))

	// 绘制大号方向指示符（居中偏上）
	AddText(canvas, indicator, w/2-30, h/2+int(fontSize/3))

	angle := 120 + rand.Float64()*180

	rotated := RotateImage(canvas, angle)
	return rotated, angle
}
