package values

import (
	"os"
	"strings"
)

func RequireValue(name string, replacements ...string) (result string) {
	if v, ok := os.LookupEnv(name); ok {
		result = v
	} else {
		errorLog.Fatalf("Error: required variable '%s' not present", name)
	}
	return replace(result, replacements)
}

func Value(name, def string, replacements ...string) (result string) {
	if v, ok := os.LookupEnv(name); ok {
		result = v
	} else {
		result = def
	}
	return replace(result, replacements)
}

func replace(value string, replacements []string) string {
	if len(replacements)%2 != 0 {
		panic("Number of replicate parameters needs to be /2.")
	}
	for i := 0; i < len(replacements); i += 2 {
		value = strings.ReplaceAll(value, "%"+replacements[i]+"%", replacements[i+1])
	}
	return value
}
