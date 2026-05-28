package snowflake

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const epoch = 1704038400000 // 2024-01-01 00:00:00 UTC in milliseconds

type Generator struct {
	mu       sync.Mutex
	workerID int64
	sequence int64
	lastTs   int64
}

var global *Generator

func init() {
	workerID := int64(1)
	if s := os.Getenv("SNOWFLAKE_WORKER_ID"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil && n >= 0 && n < 1024 {
			workerID = n
		}
	}
	global = &Generator{workerID: workerID}
}

func Next() int64 {
	return global.next()
}

func (g *Generator) next() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	ts := time.Now().UnixMilli()
	if ts < g.lastTs {
		panic(fmt.Sprintf("snowflake: clock moved backwards, refusing to generate id for %d ms", g.lastTs-ts))
	}

	if ts == g.lastTs {
		g.sequence = (g.sequence + 1) & 0xFFF
		if g.sequence == 0 {
			for ts <= g.lastTs {
				ts = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTs = ts

	return (ts-epoch)<<22 | g.workerID<<12 | g.sequence
}
