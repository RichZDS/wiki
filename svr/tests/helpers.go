// package tests contains shared test utilities for the wiki project.
package tests

import (
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a temporary directory for testing and returns a cleanup function.
// TempDir 创建测试临时目录并返回清理函数。
func TempDir(t *testing.T, pattern string) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("TempDir: %v", err)
	}
	cleanup := func() {
		os.RemoveAll(dir)
	}
	return dir, cleanup
}

// WriteFile writes content to path relative to base and returns the full path.
// WriteFile 在测试目录中写入指定内容。
func WriteFile(t *testing.T, base, rel, content string) string {
	t.Helper()
	full := filepath.Join(base, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatalf("WriteFile MkdirAll: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return full
}

// AssertNoErr fails the test if err is not nil.
// AssertNoErr 断言测试过程没有返回错误。
func AssertNoErr(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

// AssertEqual fails the test if got != want.
// AssertEqual 断言实际值与期望值相等。
func AssertEqual[T comparable](t *testing.T, got, want T, label string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", label, got, want)
	}
}

// AssertTrue fails the test if cond is false.
// AssertTrue 断言给定条件为真。
func AssertTrue(t *testing.T, cond bool, msg string) {
	t.Helper()
	if !cond {
		t.Fatalf("unexpected false: %s", msg)
	}
}
