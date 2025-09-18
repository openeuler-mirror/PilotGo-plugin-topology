// +build windows

package logging

import "errors"

func NewSyslogBackend(protocol, address, tag string) (Backend, error) {
	return nil, errors.New("Syslog backend is only supported on Linux and Unix")
}
