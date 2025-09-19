// +build !windows

package logging

import (
	"log/syslog"

	"github.com/tchap/zapext/zapsyslog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type syslogBackend struct {
	w *syslog.Writer
}

func (b *syslogBackend) Core(msgPriority zap.LevelEnablerFunc, encoder zapcore.Encoder) zapcore.Core {
	return zapsyslog.NewCore(msgPriority, encoder, b.w)
}

// NewSyslogBackend returns a new backend that outputs to syslog
func NewSyslogBackend(protocol, address, tag string) (Backend, error) {
	w, err := syslog.Dial(protocol, address, syslog.LOG_EMERG|syslog.LOG_KERN, tag)
	if err != nil {
		return nil, err
	}
	return &syslogBackend{w: w}, nil
}
