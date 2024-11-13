// internal/gateway/utils.go
package gateway

import "strings"

// split разделяет строку по разделителю и возвращает массив строк
func split(s, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}
