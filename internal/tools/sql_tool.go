package tools

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"unicode"

	"aisearch/internal/model"
	"aisearch/pkg/database"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"gorm.io/gorm"
)

const maxSQLResultRows = 200

type SQLInput = model.SQLInput
type SQLResult = model.SQLResult

// NewSQLTool 创建仅允许只读语句的 SQL 工具。
func NewSQLTool(db ...*gorm.DB) tool.InvokableTool {
	gormDB := database.DB
	if len(db) > 0 {
		gormDB = db[0]
	}

	sqlTool, err := utils.InferTool(
		"execute_sql",
		"Execute one read-only MySQL statement. Use this tool when database data is needed. "+
			"The only argument is sql, containing the complete SELECT, SHOW, DESCRIBE, DESC, or EXPLAIN statement.",
		func(ctx context.Context, input SQLInput) (SQLResult, error) {
			return executeSQL(ctx, gormDB, input.SQL)
		},
	)
	if err != nil {
		return nil
	}
	return sqlTool
}

// executeSQL 校验并执行一条只读 SQL 语句。
func executeSQL(ctx context.Context, db *gorm.DB, statement string) (SQLResult, error) {
	statement = strings.TrimSpace(statement)
	if statement == "" {
		return SQLResult{}, fmt.Errorf("sql must not be empty")
	}
	if !isReadOnlySQL(statement) {
		return SQLResult{}, fmt.Errorf("only read-only SQL is allowed (SELECT, SHOW, DESCRIBE, DESC, EXPLAIN)")
	}
	if db == nil {
		return SQLResult{}, fmt.Errorf("database is not initialized")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return SQLResult{}, fmt.Errorf("get database connection: %w", err)
	}

	rows, err := sqlDB.QueryContext(ctx, statement)
	if err != nil {
		return SQLResult{}, fmt.Errorf("execute sql: %w", err)
	}
	defer rows.Close()

	return scanSQLRows(rows)
}

// scanSQLRows 将数据库结果集转换为结构化结果。
func scanSQLRows(rows *sql.Rows) (SQLResult, error) {
	columns, err := rows.Columns()
	if err != nil {
		return SQLResult{}, fmt.Errorf("get result columns: %w", err)
	}

	result := SQLResult{
		Columns: columns,
		Rows:    make([]map[string]any, 0),
	}

	for rows.Next() {
		if result.RowCount >= maxSQLResultRows {
			result.Truncated = true
			break
		}

		values := make([]any, len(columns))
		destinations := make([]any, len(columns))
		for i := range values {
			destinations[i] = &values[i]
		}

		if err := rows.Scan(destinations...); err != nil {
			return SQLResult{}, fmt.Errorf("scan result row: %w", err)
		}

		row := make(map[string]any, len(columns))
		for i, column := range columns {
			if value, ok := values[i].([]byte); ok {
				row[column] = string(value)
			} else {
				row[column] = values[i]
			}
		}
		result.Rows = append(result.Rows, row)
		result.RowCount++
	}

	if err := rows.Err(); err != nil {
		return SQLResult{}, fmt.Errorf("read result rows: %w", err)
	}
	return result, nil
}

// isReadOnlySQL 判断语句是否属于允许执行的只读 SQL。
func isReadOnlySQL(statement string) bool {
	keyword := firstSQLKeyword(statement)
	switch keyword {
	case "SELECT", "SHOW", "DESCRIBE", "DESC", "EXPLAIN":
		return true
	default:
		return false
	}
}

// firstSQLKeyword 跳过注释并提取 SQL 的首个关键字。
func firstSQLKeyword(statement string) string {
	s := strings.TrimSpace(statement)
	for {
		switch {
		case strings.HasPrefix(s, "--"):
			if newline := strings.IndexByte(s, '\n'); newline >= 0 {
				s = strings.TrimSpace(s[newline+1:])
				continue
			}
			return ""
		case strings.HasPrefix(s, "#"):
			if newline := strings.IndexByte(s, '\n'); newline >= 0 {
				s = strings.TrimSpace(s[newline+1:])
				continue
			}
			return ""
		case strings.HasPrefix(s, "/*"):
			if end := strings.Index(s[2:], "*/"); end >= 0 {
				s = strings.TrimSpace(s[end+4:])
				continue
			}
			return ""
		}
		break
	}

	end := strings.IndexFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r)
	})
	if end < 0 {
		end = len(s)
	}
	return strings.ToUpper(s[:end])
}
