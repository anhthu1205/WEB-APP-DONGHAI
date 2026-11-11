// Package seed nhúng dữ liệu mẫu demo.sql.
package seed

import _ "embed"

// DemoSQL là nội dung seed/demo.sql.
//
//go:embed demo.sql
var DemoSQL string
