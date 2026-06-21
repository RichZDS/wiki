package job

import (
	"context"
	"testing"
	"time"
)

// TestManagerRunNow 验证任务可以被立即触发并更新运行状态。
func TestManagerRunNow(t *testing.T) {
	manager := NewManager()
	done := make(chan struct{})

	if err := manager.Register("demo", time.Hour, func(context.Context) error {
		close(done)
		return nil
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	snapshot, err := manager.RunNow("demo")
	if err != nil {
		t.Fatalf("run now: %v", err)
	}
	if !snapshot.Running {
		t.Fatal("expected job to be running")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("job did not run")
	}
	manager.Wait()

	snapshot, err = manager.Get("demo")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if snapshot.LastStatus != jobStatusSucceeded {
		t.Fatalf("LastStatus got %q, want %q", snapshot.LastStatus, jobStatusSucceeded)
	}
}

// TestManagerDisableCancelsRunningJob 验证停用任务会取消正在运行的执行。
func TestManagerDisableCancelsRunningJob(t *testing.T) {
	manager := NewManager()
	started := make(chan struct{})
	canceled := make(chan struct{})

	if err := manager.Register("slow", time.Hour, func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		close(canceled)
		return ctx.Err()
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	if _, err := manager.RunNow("slow"); err != nil {
		t.Fatalf("run now: %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("job did not start")
	}

	snapshot, err := manager.Disable("slow")
	if err != nil {
		t.Fatalf("disable: %v", err)
	}
	if snapshot.Enabled {
		t.Fatal("expected job to be disabled")
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("job was not canceled")
	}
	manager.Wait()

	snapshot, err = manager.Get("slow")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if snapshot.LastStatus != jobStatusCanceled {
		t.Fatalf("LastStatus got %q, want %q", snapshot.LastStatus, jobStatusCanceled)
	}
}

// TestShouldPersistLevel 验证日志入库级别阈值判断。
func TestShouldPersistLevel(t *testing.T) {
	if shouldPersistLevel(jobLogDebug, jobLogInfo) {
		t.Fatal("debug should not pass info threshold")
	}
	if !shouldPersistLevel(jobLogError, jobLogWarn) {
		t.Fatal("error should pass warn threshold")
	}
}
