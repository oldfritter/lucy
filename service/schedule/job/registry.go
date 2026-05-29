package job

import "github.com/robfig/cron/v3"

// Job 定时任务定义
type Job struct {
	Name string          // 任务名称（日志标识）
	Spec string          // cron 表达式，支持 @every 1h 等语法
	Func func()          // 任务执行函数
}

// Registry 全局任务注册表，在 init() 中注册
var Registry []Job

// Register 向注册表添加任务
func Register(j Job) {
	// 默认使用标准 cron 解析器（含秒字段）
	if _, err := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor).Parse(j.Spec); err != nil {
		panic("invalid cron spec for job " + j.Name + ": " + err.Error())
	}
	Registry = append(Registry, j)
}
