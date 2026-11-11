// Package migrations nhúng các file SQL migration (chạy theo thứ tự tên file).
package migrations

import "embed"

// FS chứa các file *.sql.
//
//go:embed *.sql
var FS embed.FS
