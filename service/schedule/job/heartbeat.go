package job

import "log"

func init() {
	Register(Job{
		Name: "heartbeat",
		Spec: "@every 5m",
		Func: heartbeat,
	})
}

// heartbeat 心跳任务：证明调度器正常运行
func heartbeat() {
	log.Println("[heartbeat] scheduler is alive")
}
