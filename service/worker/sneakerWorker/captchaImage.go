package sneakerWorker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"time"

	sneaker "github.com/oldfritter/sneaker-go/v3"

	"github.com/oldfritter/lucy/base"
	captchaImage "github.com/oldfritter/lucy/lib/captcha"
	"github.com/oldfritter/lucy/lib/storage/oss"
)

const defaultBackground = "config/background/c71eda17095e9a92e300ca207f09c778.jpg"

// CaptchaPayload 验证码图片生成任务载荷
type CaptchaPayload struct {
	Prompts       []string `json:"prompts"`
	Key           string   `json:"key"`
	BackgroundURL string   `json:"backgroundUrl,omitempty"` // 可选：投放的自定义背景图 URL
}

// CaptchaImageWorker 验证码图片生成 Worker
type CaptchaImageWorker struct {
	sneaker.Worker
}

// CaptchaImageWorkerInstance 导出供 main.go 直接使用
var CaptchaImageWorkerInstance = &CaptchaImageWorker{
	Worker: sneaker.Worker{
		Name:    "CaptchaImage",
		Threads: 1,
		Durable: true,
	},
}

// NewCaptchaImageWorker 根据配置创建 CaptchaImageWorker 实例
func NewCaptchaImageWorker(cfg base.WorkerConfig) *CaptchaImageWorker {
	return &CaptchaImageWorker{
		Worker: sneaker.Worker{
			Name:    cfg.Name,
			Queue:   cfg.Queue,
			Log:     cfg.Log,
			Threads: cfg.Threads,
			Durable: cfg.Durable,
		},
	}
}

// Work 实现 sneaker.Worker 接口的 Work 方法
func (w *CaptchaImageWorker) Work(payloadJson *[]byte) (err error) {
	start := time.Now().UnixNano()

	var payload CaptchaPayload
	if err = json.Unmarshal(*payloadJson, &payload); err != nil {
		w.LogError("payload unmarshal failed: ", err, ", time:", (time.Now().UnixNano()-start)/1000000, " ms")
		return
	}

	w.LogInfo("start processing, prompts: ", payload.Prompts, ", key: ", payload.Key)

	bg := resolveBackground(payload.BackgroundURL)
	img := captchaImage.GenerateCaptchaImage(payload.Prompts, bg)

	var buf bytes.Buffer
	if err = png.Encode(&buf, img); err != nil {
		w.LogError("png encode failed: ", err, ", time:", (time.Now().UnixNano()-start)/1000000, " ms")
		return
	}

	b := buf.Bytes()
	if _, err = oss.PutObject(payload.Key, &b); err != nil {
		w.LogError("oss upload failed: ", err, ", time:", (time.Now().UnixNano()-start)/1000000, " ms")
		return
	}

	w.LogInfo("completed, key: ", payload.Key, ", time:", (time.Now().UnixNano()-start)/1000000, " ms")
	return
}

// resolveBackground 优先使用远程背景，失败则回退默认
func resolveBackground(url string) string {
	if url == "" {
		return defaultBackground
	}
	resp, err := http.Get(url)
	if err != nil {
		return defaultBackground
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return defaultBackground
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return defaultBackground
	}
	// 验证是否为有效图片
	if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
		return defaultBackground
	}
	// 写入临时文件（纳秒时间戳防并发冲突）
	tmpPath := fmt.Sprintf("config/background/campaign_%d.jpg", time.Now().UnixNano())
	if err := writeFile(tmpPath, data); err != nil {
		return defaultBackground
	}
	return tmpPath
}

func writeFile(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}
