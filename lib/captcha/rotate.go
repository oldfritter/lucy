package captcha

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
)

// RotateImage 将图片按指定角度旋转（逆时针为正），返回旋转后图像
func RotateImage(src image.Image, angle float64) image.Image {
	bounds := src.Bounds()
	rad := angle * math.Pi / 180.0

	w, h := float64(bounds.Dx()), float64(bounds.Dy())
	cos, sin := math.Abs(math.Cos(rad)), math.Abs(math.Sin(rad))
	nw := int(w*cos + h*sin)
	nh := int(w*sin + h*cos)

	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	cx, cy := w/2, h/2
	ncx, ncy := float64(nw)/2, float64(nh)/2

	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			nx := float64(x) - ncx
			ny := float64(y) - ncy

			// 逆向旋转矩阵：从目标坐标反推源坐标
			sx := nx*math.Cos(rad) + ny*math.Sin(rad) + cx
			sy := -nx*math.Sin(rad) + ny*math.Cos(rad) + cy

			sxInt, syInt := int(sx+0.5), int(sy+0.5)
			if sxInt >= 0 && sxInt < bounds.Dx() && syInt >= 0 && syInt < bounds.Dy() {
				dst.Set(x, y, src.At(sxInt, syInt))
			}
		}
	}

	return dst
}

// GenerateRotateCaptcha 生成旋转验证码图片。
// backgroundPath: 背景图片路径
// indicator: 方向指示文字，如 "▲"
// 返回：旋转后的图片、旋转角度（0-360）
func GenerateRotateCaptcha(backgroundPath, indicator string) (image.Image, float64) {
	img := LoadBackground(backgroundPath)
	bounds := img.Bounds()

	// 创建画布，绘制指示标志
	canvas := image.NewRGBA(bounds)
	draw.Draw(canvas, bounds, img, bounds.Min, draw.Src)

	// 在图片中央偏上绘制大号方向指示符
	AddText(canvas, indicator, bounds.Dx()/2-30, bounds.Dy()/2+20)

	// 随机角度：120° ~ 300°，避开 0°（朝上）附近 ±60°
	angle := 120 + rand.Float64()*180

	rotated := RotateImage(canvas, angle)
	return rotated, angle
}
