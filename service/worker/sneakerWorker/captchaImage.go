package sneakerWorker

import (
	"bytes"
	"encoding/json"
	"image/png"
	"time"

	sneaker "github.com/oldfritter/sneaker-go/v3"

	"github.com/oldfritter/lucy/base"
	captchaImage "github.com/oldfritter/lucy/lib/captcha"
	"github.com/oldfritter/lucy/lib/storage/oss"
)

const defaultBackground = "config/background/c71eda17095e9a92e300ca207f09c778.jpg"

// CaptchaPayload 验证码图片生成任务载荷
type CaptchaPayload struct {
	Prompts []string `json:"prompts"`
	Key     string   `json:"key"`
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

	img := captchaImage.GenerateCaptchaImage(payload.Prompts, defaultBackground)

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
