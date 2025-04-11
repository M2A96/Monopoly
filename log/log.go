package log

import (
	"fmt"
	"runtime"
	"strings"

	"go.uber.org/zap"
)

// GetFileLine is a function.
func GetFileLine() zap.Field {
	_, filename, line, ok := runtime.Caller(2)
	if !ok {
		return zap.Skip()
	}

	filenameSplit := strings.SplitN(filename, "@", 2)
	if 2 <= len(filenameSplit) {
		filename = filenameSplit[1]
	}

	return zap.String("source", fmt.Sprintf(`%v:%v`, filename, line))
}
