package consts

import "time"

const (
	// ModelHealthInterval 模型健康检查的默认执行间隔。
	ModelHealthInterval = 5 * time.Minute
	// ModelProbeTimeout 单次模型探测的超时时间。
	ModelProbeTimeout = 15 * time.Second
	// ModelFailReasonNotFound 模型健康检查失败原因：模型未找到。
	ModelFailReasonNotFound = "model not found"
)
