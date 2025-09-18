/*
 * Copyright (c) KylinSoft  Co., Ltd. 2025.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2.
 * See LICENSE file for more details.
 * Author: waneng <wanfeng@kylinos.cn>
 */
package config

import (
	"fmt"
	"os"

	"github.com/skydive-project/skydive/graffiti/logging"
)

// InitLogging set up logging based on the section "logging" of
// the configuration file
func InitLogging() error {
	color := GetBool("logging.color")
	id := GetString("host_id") + ":" + GetString("logging.id")
	defaultEncoder := cfg.GetString("logging.encoder")
	defaultLogLevel := cfg.GetString("logging.level")

	var err error
	var backend logging.Backend
	var loggers []*logging.LoggerConfig
	for _, name := range cfg.GetStringSlice("logging.backends") {
		switch name {
		case "file":
			filename := cfg.GetString("logging.file.path")
			backend, err = logging.NewFileBackend(filename)
			if err != nil {
				return err
			}
		case "syslog":
			syslogTag := cfg.GetString("logging.syslog.tag")
			protocol := cfg.GetString("logging.syslog.protocol")
			addr := cfg.GetString("logging.syslog.address")
			backend, err = logging.NewSyslogBackend(protocol, addr, syslogTag)
			if err != nil {
				return err
			}
		case "stderr":
			backend = logging.NewStdioBackend(os.Stderr)
		case "stdout":
			backend = logging.NewStdioBackend(os.Stdout)
		default:
			return fmt.Errorf("Invalid logging backend: %s", name)
		}

		prefix := "logging." + name
		encoder := defaultEncoder
		logLevel := defaultLogLevel
		if e := cfg.GetString(prefix + ".encoder"); e != "" {
			encoder = e
		}
		if l := cfg.GetString(prefix + ".level"); l != "" {
			logLevel = l
		}

		loggers = append(loggers, logging.NewLoggerConfig(backend, logLevel, encoder))
	}

	return logging.InitLogging(id, color, loggers)
}
