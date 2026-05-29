package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/oldfritter/lucy/initialize"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/lib/kv"
	_ "github.com/oldfritter/lucy/lib/storage/oss"
	"github.com/oldfritter/lucy/service/schedule/job"
	"github.com/oldfritter/lucy/util"
)

func main() {
	defer kv.CloseRedisPools()

	initialize.MigrateDB()
	defer db.CloseDB()

	log.SetOutput(util.GetLogFile("schedule"))

	c := cron.New(cron.WithSeconds(), cron.WithLocation(time.Local))

	log.Println("Schedule service starting ...")
	log.Printf("Registered %d job(s)", len(job.Registry))
	for _, j := range job.Registry {
		id, err := c.AddFunc(j.Spec, j.Func)
		if err != nil {
			log.Printf("register job %q failed: %v", j.Name, err)
			continue
		}
		log.Printf("  [%d] %s  spec=%s", id, j.Name, j.Spec)
	}

	c.Start()
	log.Println("Scheduler started")

	setPid()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down scheduler ...")
	ctx := c.Stop()
	<-ctx.Done()
	log.Println("All jobs finished")
}

func setPid() {
	if err := os.MkdirAll("pids", 0755); err != nil {
		log.Fatalf("create folder error: %v", err)
	}
	if err := os.WriteFile("pids/schedule.pid", []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		log.Fatalf("write pid error: %v", err)
	}
}
