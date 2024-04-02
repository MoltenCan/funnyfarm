package cdb

import (
	"log/slog"
	"runtime"
	"strings"
	"time"
)

// a really crappy beancounter
// just debug logs out the times
// uses the runtime to get the function name

type beanCounter struct {
	startTime time.Time
	f         string
}

func BCStart() beanCounter {
	// get a nice func name
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	pcF, _ := runtime.CallersFrames([]uintptr{pcs[0]}).Next()
	parts := strings.Split(pcF.Function, "/")

	return beanCounter{
		startTime: time.Now(),
		f:         parts[len(parts)-1],
	}
}

func BCCount(c beanCounter) {
	slog.Debug("beancounter", "time", time.Since(c.startTime), "function", c.f)
}
