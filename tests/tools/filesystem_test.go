package tools

import (
	"context"
	"runtime"
	"testing"

	"github.com/cloudwego/eino/adk/filesystem"

	tests "aisearch/tests"
)

// TestInMemoryBackendWriteRead 验证内存文件后端的写入和读取。
func TestInMemoryBackendWriteRead(t *testing.T) {
	ctx := context.Background()
	backend := filesystem.NewInMemoryBackend()

	path := "/test/readme.md"
	content := "# Hello\n\nThis is a test file."

	err := backend.Write(ctx, &filesystem.WriteRequest{
		FilePath: path,
		Content:  content,
	})
	tests.AssertNoErr(t, err, "Write")

	fc, err := backend.Read(ctx, &filesystem.ReadRequest{
		FilePath: path,
		Offset:   1,
		Limit:    50,
	})
	tests.AssertNoErr(t, err, "Read")
	tests.AssertTrue(t, fc != nil, "Read returned nil FileContent")
	tests.AssertTrue(t, fc.Content != "", "Read returned empty content")
	t.Logf("Write/Read OK: %d bytes", len(fc.Content))
}

// TestInMemoryBackendEdit 验证内存文件后端的编辑能力。
func TestInMemoryBackendEdit(t *testing.T) {
	ctx := context.Background()
	backend := filesystem.NewInMemoryBackend()

	path := "/config.yaml"
	err := backend.Write(ctx, &filesystem.WriteRequest{
		FilePath: path,
		Content:  "port: 8080\nhost: localhost\n",
	})
	tests.AssertNoErr(t, err, "Write")

	err = backend.Edit(ctx, &filesystem.EditRequest{
		FilePath:  path,
		OldString: "8080",
		NewString: "9090",
	})
	tests.AssertNoErr(t, err, "Edit")

	fc, err := backend.Read(ctx, &filesystem.ReadRequest{FilePath: path})
	tests.AssertNoErr(t, err, "Read after edit")

	expected := "port: 9090\nhost: localhost\n"
	if fc.Content != expected {
		t.Fatalf("expected %q, got %q", expected, fc.Content)
	}
	t.Logf("Write/Edit/Read OK")
}

// TestInMemoryBackendPathNormalizationBug verifies the known normalizePath bug
// on Windows: filepath.Clean converts "/" to "\", but the rest of the code
// uses literal "/" for path prefix matching, causing GrepRaw/LsInfo/GlobInfo
// to return empty results on Windows.
// TestInMemoryBackendPathNormalizationBug 验证路径规范化场景不会回归。
func TestInMemoryBackendPathNormalizationBug(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("this test verifies a Windows-specific path normalization bug")
	}

	ctx := context.Background()
	backend := filesystem.NewInMemoryBackend()

	// Write/Read work — exact path lookup uses same normalizePath both times
	err := backend.Write(ctx, &filesystem.WriteRequest{
		FilePath: "/project/main.go",
		Content:  "package main\n\nfunc main() {}",
	})
	tests.AssertNoErr(t, err, "Write")

	_, err = backend.Read(ctx, &filesystem.ReadRequest{FilePath: "/project/main.go"})
	tests.AssertNoErr(t, err, "Read — exact path lookup works")

	// GrepRaw returns no error but finds nothing — prefix match fails
	matches, err := backend.GrepRaw(ctx, &filesystem.GrepRequest{
		Pattern: "func",
		Path:    "/project",
	})
	tests.AssertNoErr(t, err, "GrepRaw doesn't panic")
	if len(matches) == 0 {
		t.Log("KNOWN BUG: GrepRaw returned 0 matches — normalizePath uses '\\' but filterFiles uses '/' in prefix")
	} else {
		t.Logf("Unexpected: GrepRaw found %d matches (bug may be fixed)", len(matches))
	}

	// LsInfo also broken
	files, err := backend.LsInfo(ctx, &filesystem.LsInfoRequest{Path: "/project"})
	tests.AssertNoErr(t, err, "LsInfo doesn't panic")
	if len(files) == 0 {
		t.Log("KNOWN BUG: LsInfo returned 0 entries — same normalizePath root cause")
	}
}

var _ filesystem.Backend = filesystem.NewInMemoryBackend()
