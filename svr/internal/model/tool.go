package model

type WeatherInput struct {
	City string `json:"city" jsonschema:"description=城市名称（中文），例如 深圳、北京"`
	Date string `json:"date" jsonschema:"description=查询日期，支持 today/tomorrow/YYYY-MM-DD"`
}

type WeatherOutput struct {
	City        string `json:"city"`
	Date        string `json:"date"`
	Temperature int    `json:"temperature"`
	Condition   string `json:"condition"`
	Humidity    int    `json:"humidity"`
}

type SQLInput struct {
	SQL string `json:"sql" jsonschema:"description=需要执行的完整只读 MySQL 语句"`
}

type SQLResult struct {
	Columns   []string         `json:"columns"`
	Rows      []map[string]any `json:"rows"`
	RowCount  int              `json:"row_count"`
	Truncated bool             `json:"truncated"`
}
