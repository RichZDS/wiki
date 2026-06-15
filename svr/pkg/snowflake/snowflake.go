package snowflake

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"wiki/internal/model"
	"wiki/internal/model/consts"
)

type Generator = model.SnowflakeGenerator

var global *model.SnowflakeGenerator

// init 根据环境变量初始化全局雪花编号生成器。
func init() {
	workerID := int64(1)
	if s := os.Getenv("SNOWFLAKE_WORKER_ID"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil && n >= 0 && n < 1024 {
			workerID = n
		}
	}
	global = &model.SnowflakeGenerator{WorkerID: workerID}
}

// Next 生成下一个全局唯一的雪花编号。
func Next() int64 {
	return next(global)
}

// next 根据时间戳、工作节点和序列号生成雪花编号。
func next(g *model.SnowflakeGenerator) int64 {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	ts := time.Now().UnixMilli()
	if ts < g.LastTs {
		panic(fmt.Sprintf("snowflake: clock moved backwards, refusing to generate id for %d ms", g.LastTs-ts))
	}

	if ts == g.LastTs {
		g.Sequence = (g.Sequence + 1) & 0xFFF
		if g.Sequence == 0 {
			for ts <= g.LastTs {
				ts = time.Now().UnixMilli()
			}
		}
	} else {
		g.Sequence = 0
	}

	g.LastTs = ts

	return (ts-consts.Epoch)<<22 | g.WorkerID<<12 | g.Sequence
}
