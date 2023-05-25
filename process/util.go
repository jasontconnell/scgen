package process

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

func getUnderscoreUppercaseName(name string) string {
	name = strings.Replace(strings.Title(name), " ", "_", -1)
	return getCleanName(name)
}

func getUnderscoreLowercaseName(name string) string {
	return strings.ToLower(getUnderscoreUppercaseName(name))
}

func getCleanNameFunc(setting string) func(string) string {
	var ret func(string) string
	switch strings.ToLower(strings.ReplaceAll(setting, "_", "")) {
	case "", "pascalcase":
		ret = getCleanName
	case "pascalcaseunderscore":
		ret = getUnderscoreUppercaseName
	case "lowercaseunderscore":
		ret = getUnderscoreLowercaseName
	default:
		panic("Name style not recognized: " + setting)
	}
	return ret
}
