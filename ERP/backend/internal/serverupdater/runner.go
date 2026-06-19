package serverupdater

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

const Version = "cbmp-server-updater/1.0.0"

type Runner struct {
	cfg       Config
	client    *Client
	installer Installer
	logger    *log.Logger
}

func NewRunner(cfg Config, logger *log.Logger) (*Runner, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = log.Default()
	}
	return &Runner{cfg: cfg, client: client, installer: NewInstaller(cfg), logger: logger}, nil
}

func (r *Runner) Run(ctx context.Context) error {
	interval := time.Duration(r.cfg.PollIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	for {
		if err := r.RunOnce(ctx); err != nil {
			r.logger.Printf("server updater run failed: %v", err)
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (r *Runner) RunOnce(ctx context.Context) error {
	poll, err := r.client.Poll(ctx)
	if err != nil {
		return err
	}
	for _, task := range poll.Tasks {
		if !r.shouldHandle(task) {
			continue
		}
		if err := r.handleTask(ctx, task); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) shouldHandle(task ProductSystemUpdateTask) bool {
	if r.cfg.TargetComponent != "" && !strings.EqualFold(strings.TrimSpace(task.Component), r.cfg.TargetComponent) {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(task.Status)) {
	case "queued", "assigned", "running":
		return true
	default:
		return false
	}
}

func (r *Runner) handleTask(ctx context.Context, task ProductSystemUpdateTask) error {
	r.logger.Printf("handling server update task %s component=%s action=%s target=%s", task.TaskNo, task.Component, task.Action, task.Version)
	_ = r.client.Report(ctx, task.TaskNo, ReportRequest{Status: "running", Progress: 10, Step: "accepted", Message: "服务端更新器已接收任务"})
	var releaseDir string
	var err error
	switch strings.ToLower(strings.TrimSpace(task.Action)) {
	case "rollback":
		_ = r.client.Report(ctx, task.TaskNo, ReportRequest{Status: "running", Progress: 35, Step: "rollback_prepare", Message: "服务端更新器正在准备回滚"})
		releaseDir, err = r.installer.Rollback(ctx, task)
	default:
		_ = r.client.Report(ctx, task.TaskNo, ReportRequest{Status: "running", Progress: 25, Step: "download", Message: "服务端更新器正在下载更新包"})
		download, artifact, downloadErr := r.client.DownloadTask(ctx, task)
		if downloadErr != nil {
			err = downloadErr
			break
		}
		_ = r.client.Report(ctx, task.TaskNo, ReportRequest{Status: "running", Progress: 55, Step: "install", Message: "服务端更新器正在写入蓝绿发布目录"})
		releaseDir, err = r.installer.Apply(ctx, task, download, artifact)
	}
	if err != nil {
		if r.cfg.AutoRollbackOnFailure && !strings.EqualFold(task.Action, "rollback") {
			_ = r.client.Report(ctx, task.TaskNo, ReportRequest{Status: "running", Progress: 85, Step: "auto_rollback", Message: "服务端更新器执行失败，正在自动回滚上一版本", Error: err.Error()})
			rollbackTask := task
			rollbackTask.Action = "rollback"
			rollbackDir, rollbackErr := r.installer.Rollback(ctx, rollbackTask)
			if rollbackErr == nil {
				currentVersion := fallbackText(task.FromVersion, "previous")
				message := "服务端更新器更新失败，已自动回滚上一版本，发布目录 " + rollbackDir
				reportErr := r.client.Report(ctx, task.TaskNo, ReportRequest{Status: "rolled_back", Progress: 100, Step: "auto_rollback", Message: message, Error: err.Error(), CurrentVersion: currentVersion})
				if reportErr != nil {
					return fmt.Errorf("%w; auto rollback succeeded but report failed: %v", err, reportErr)
				}
				r.logger.Printf("task %s failed and auto rolled back to %s release=%s", task.TaskNo, currentVersion, rollbackDir)
				return nil
			}
			err = fmt.Errorf("%w; auto rollback failed: %v", err, rollbackErr)
		}
		reportErr := r.client.Report(ctx, task.TaskNo, ReportRequest{Status: "failed", Progress: 100, Step: "failed", Message: "服务端更新器执行失败", Error: err.Error()})
		if reportErr != nil {
			return fmt.Errorf("%w; report failed: %v", err, reportErr)
		}
		return err
	}
	status := "succeeded"
	currentVersion := task.Version
	message := "服务端更新器已完成更新并通过本地健康检查"
	if strings.EqualFold(task.Action, "rollback") {
		status = "rolled_back"
		currentVersion = task.FromVersion
		message = "服务端更新器已完成回滚并通过本地健康检查"
	}
	err = r.client.Report(ctx, task.TaskNo, ReportRequest{Status: status, Progress: 100, Step: "health_check", Message: message + "，发布目录 " + releaseDir, CurrentVersion: currentVersion})
	if err == nil {
		r.logger.Printf("task %s completed version=%s release=%s", task.TaskNo, currentVersion, releaseDir)
	}
	return err
}
