package values

import (
	"os"
	"strings"
)

func Value(name, def string, replacements ...string) (result string) {
	if v, ok := os.LookupEnv(name); ok {
		result = v
	} else {
		result = def
	}
	if len(replacements)%2 != 0 {
		panic("Number of replicate parameters needs to be /2.")
	}
	for i := 0; i < len(replacements); i += 2 {
		result = strings.ReplaceAll(result, "%"+replacements[i]+"%", replacements[i+1])
	}
	return
}
