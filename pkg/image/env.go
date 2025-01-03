package image

import (
	"os"
	"strings"
)

// struct EnvVar {
// 	Name string
// 	Value string
// }

func EnvVariables() map[string]string {
	env := map[string]string{}

	for _, item := range os.Environ() {
		splits := strings.SplitN(item, "=", 2)
		name := splits[0]
		val := splits[1]
		env[name] = val
	}

	return env
}
