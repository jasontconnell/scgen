package processor

import (
	"strings"
)

func getCleanName(name string) string {
	prefix := ""
	if strings.IndexAny(name, "0123456789") == 0 {
		prefix = "_"
	}
	return prefix + strings.Replace(strings.Replace(strings.Title(name), "-", "", -1), " ", "", -1)
}