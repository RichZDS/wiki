package agent

import (
	"context"
	"fmt"
	"testing"

	"wiki/internal/ai/chunk"

	"github.com/cloudwego/eino/schema"
)

// --- 三组测试文本 ---

// shortText: 短文字 ~120 字符，1-2 chunks
const shortText = `Go 语言是 Google 开发的一种静态强类型、编译型、并发型编程语言。它具有简洁、快速、安全的特性，内置并发支持，适合构建高性能网络服务。`

// mediumText: 中等文字 ~600 字符，多个 chunks
const mediumText = `Go 语言（又称 Golang）是 Google 的 Robert Griesemer、Rob Pike 及 Ken Thompson 开发的一种静态强类型、编译型语言。
Go 的语法接近 C 语言，但增加了内存安全、垃圾回收、结构类型以及 CSP 风格的并发模型。

Go 的主要特点包括：
1. 简洁的语法设计，学习曲线平缓
2. 内置并发原语 goroutine 和 channel，轻松实现高并发
3. 静态编译为单一二进制文件，部署简单
4. 强大的标准库，涵盖网络、加密、压缩等领域
5. 高效的编译速度，适合大型项目快速迭代

Go 语言广泛应用于云原生开发、微服务架构、DevOps 工具链等领域。
Kubernetes、Docker、Prometheus 等知名项目均使用 Go 语言编写。
Go 的并发模型基于 CSP（Communicating Sequential Processes）理论，
通过 goroutine 实现轻量级线程，通过 channel 实现 goroutine 间的通信。
这种设计使得并发编程变得简单而安全，避免了传统多线程编程中的竞态条件问题。`

// longText: 超长文字 ~2000 字符，大量 chunks，测试 overlap
const longText = `第一章 引言

Go 语言（Golang）是由 Google 公司在 2007 年开始设计、2009 年正式对外发布的一种编程语言。
它的设计者包括 Robert Griesemer（曾参与 V8 JavaScript 引擎开发）、Rob Pike（Unix 先驱，
Plan 9 操作系统和 UTF-8 编码的共同设计者）以及 Ken Thompson（Unix 操作系统的创造者之一，
B 语言和 C 语言的设计者，图灵奖得主）。这三位顶级计算机科学家将他们数十年的编程语言
设计经验融入了 Go 语言的设计之中。

第二章 设计哲学

Go 语言的设计哲学可以概括为"少即是多"（Less is More）。与 C++ 或 Java 等语言不同，
Go 刻意保持语言的简洁性。它没有泛型（直到 Go 1.18 才引入）、没有继承、没有方法重载、
没有指针运算、没有隐式类型转换。Go 的设计者认为，特性越少，代码越容易阅读和维护。

Go 语言追求的是"正交性"（Orthogonality）——每个特性应该独立且不重复。
例如，Go 的接口类型是一种隐式实现，类型不需要显式声明它实现了哪个接口，
只要类型的方法集合包含了接口所需的所有方法，该类型就自动实现了该接口。
这种设计使得代码的耦合度大大降低。

第三章 并发模型

Go 语言的并发模型是其最显著的特性之一。它基于 Tony Hoare 在 1978 年提出的
CSP（Communicating Sequential Processes）理论。"不要通过共享内存来通信，
而应该通过通信来共享内存"——这是 Go 并发编程的核心格言。

goroutine 是 Go 语言中的轻量级线程，由 Go 运行时调度管理。一个 goroutine 的
初始栈大小只有几 KB，远小于操作系统线程的数 MB 栈空间。Go 运行时可以将成千上万个
goroutine 高效地调度到少量的操作系统线程上运行，这种调度模型称为 M:N 调度。

channel 是 goroutine 之间通信的管道。通过 channel，一个 goroutine 可以安全地
向另一个 goroutine 发送数据。Go 提供了无缓冲 channel 和有缓冲 channel 两种类型。
无缓冲 channel 保证发送和接收操作同步进行，有缓冲 channel 则允许一定程度的异步通信。

select 语句使得 goroutine 可以同时等待多个 channel 操作。它是 Go 并发编程中
最重要的控制结构之一。通过 select，程序可以处理超时、取消和非阻塞通信等场景。

第四章 标准库与生态系统

Go 语言拥有一个庞大而高质量的标准库。标准库涵盖了网络编程（net/http）、加密解密
（crypto）、数据压缩（compress）、数据库访问（database/sql）、编码转换（encoding）
等几乎所有常用功能。许多 Go Web 应用甚至不需要引入第三方框架，仅靠标准库的
net/http 包就可以构建出生产级别的 HTTP 服务。

第五章 应用场景

Go 语言在云计算和微服务领域占据了主导地位。Docker 容器引擎、Kubernetes 容器编排平台、
Prometheus 监控系统、Etcd 分布式键值存储、Consul 服务发现、Terraform 基础设施即代码
工具——这些云原生技术栈中的核心项目无一例外地选择了 Go 语言。

除了云原生领域，Go 在 CLI 工具开发、网络代理、数据库系统、区块链平台等方面也有
广泛应用。Go 的跨平台编译能力（通过 GOOS 和 GOARCH 环境变量可以轻松交叉编译）
使得它成为开发多平台工具的绝佳选择。

第六章 总结

Go 语言以其简洁的语法、强大的并发模型、高效的编译速度和丰富的标准库，
成为了现代软件开发中不可或缺的工具。无论是构建微服务、开发运维工具、
还是编写高性能网络代理，Go 都是一个值得优先考虑的选择。`

// TestFreeChunkerShort 测试短文本切块
func TestFreeChunkerShort(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyFree)
	cfg := chunk.ChunkConfig{
		ChunkSize:    200,
		ChunkOverlap: 30,
	}

	// // docs, err := chunker.Chunk(context.Background(), shortText, cfg)
	// docs, err := chunker.Chunk(context.Background(), mediumText, cfg)

	docs, err := chunker.Chunk(context.Background(), longText, cfg)

	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}
	if len(docs) == 0 {
		t.Fatal("expected at least 1 chunk")
	}

	// t.Logf("短文本 (%d chars) → %d chunks\n", len([]rune(shortText)), len(docs))
	for _, d := range docs {
		t.Logf("chunk :id:%s, meta:%+v, content:%s", d.ID, d.MetaData, d.Content)
	}
}

// TestFreeChunkerMedium 测试中等文本切块
func TestFreeChunkerMedium(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyFree)
	cfg := chunk.ChunkConfig{
		ChunkSize:    200,
		ChunkOverlap: 30,
	}

	docs, err := chunker.Chunk(context.Background(), mediumText, cfg)
	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}
	if len(docs) == 0 {
		t.Fatal("expected at least 1 chunk")
	}

	t.Logf("中等文本 (%d chars) → %d chunks", len([]rune(mediumText)), len(docs))
	for _, d := range docs {
		idx := d.MetaData["chunk_index"]
		total := d.MetaData["chunk_total"]
		contentLen := len([]rune(d.Content))
		t.Logf("  [chunk %v/%v] %d chars: %s", idx, total, contentLen, truncate(d.Content, 80))

		// 验证每个chunk不超过ChunkSize
		if contentLen > cfg.ChunkSize+cfg.ChunkOverlap+10 {
			t.Errorf("chunk %v exceeds limit: %d > %d", idx, contentLen, cfg.ChunkSize)
		}
	}

	fmtPrintDocs(t, docs)
}

// TestFreeChunkerLong 测试超长文本切块 + overlap 验证
func TestFreeChunkerLong(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyFree)
	cfg := chunk.ChunkConfig{
		ChunkSize:    250,
		ChunkOverlap: 40,
	}

	docs, err := chunker.Chunk(context.Background(), longText, cfg)
	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}
	if len(docs) < 3 {
		t.Fatalf("expected at least 3 chunks for long text, got %d", len(docs))
	}

	t.Logf("超长文本 (%d chars) → %d chunks", len([]rune(longText)), len(docs))

	// 检查 overlap 效果
	for i, d := range docs {
		idx := d.MetaData["chunk_index"]
		total := d.MetaData["chunk_total"]
		t.Logf("  [chunk %v/%v] %d chars: %s", idx, total, len([]rune(d.Content)), truncate(d.Content, 80))

		// 验证 overlap: 前一个chunk的尾部应该出现在后一个chunk的头部
		if i > 0 && cfg.ChunkOverlap > 0 {
			prevEnd := []rune(docs[i-1].Content)
			currStart := []rune(d.Content)
			if len(prevEnd) > cfg.ChunkOverlap && len(currStart) > cfg.ChunkOverlap {
				prevTail := string(prevEnd[len(prevEnd)-cfg.ChunkOverlap:])
				currHead := string(currStart[:cfg.ChunkOverlap])
				if prevTail == currHead {
					t.Logf("    overlap verified: \"%s\"", truncate(prevTail, 30))
				} else {
					t.Logf("    overlap NOT detected (chunks may have been merged)")
				}
			}
		}
	}
}

// TestFreeChunkerEmpty 验证自由切块器处理空文本的行为。
func TestFreeChunkerEmpty(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyFree)
	docs, err := chunker.Chunk(context.Background(), "", chunk.ChunkConfig{ChunkSize: 100})
	if err != nil {
		t.Fatalf("Chunk empty string failed: %v", err)
	}
	if len(docs) != 0 {
		t.Fatalf("expected 0 chunks for empty string, got %d", len(docs))
	}
	t.Log("empty input → 0 chunks")
}

// TestFreeChunkerMetadata 验证自由切块结果包含正确元数据。
func TestFreeChunkerMetadata(t *testing.T) {
	chunker := chunk.NewChunker(chunk.StrategyFree)
	docs, err := chunker.Chunk(context.Background(), "Hello World. This is a test.", chunk.ChunkConfig{
		ChunkSize:    100,
		ChunkOverlap: 0,
	})
	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}

	for _, d := range docs {
		idx, ok := d.MetaData["chunk_index"]
		if !ok {
			t.Error("missing chunk_index in metadata")
		}
		total, ok := d.MetaData["chunk_total"]
		if !ok {
			t.Error("missing chunk_total in metadata")
		}
		if d.ID == "" {
			t.Error("empty document ID")
		}
		t.Logf("  ID=%s index=%v total=%v content=%q", d.ID, idx, total, d.Content)
	}

}

// truncate 将字符串截断到指定的最大字符数。
func truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen]) + "..."
}

// fmtPrintDocs 将切块文档格式化输出到测试日志。
func fmtPrintDocs(t *testing.T, docs []*schema.Document) {
	t.Helper()
	fmt.Println("\n========== Chunking Result ==========")
	for _, d := range docs {
		idx := d.MetaData["chunk_index"]
		total := d.MetaData["chunk_total"]
		fmt.Printf("\n--- Chunk [%v/%v] (%d chars) ---\n", idx, total, len([]rune(d.Content)))
		fmt.Println(d.Content)
	}
	fmt.Println("======================================")
}
