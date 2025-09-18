/*
 * Copyright (c) KylinSoft  Co., Ltd. 2025.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2.
 * See LICENSE file for more details.
 * Author: waneng <wanfeng@kylinos.cn>
 */
package config

import (
	"net/http"
	"net/url"
	"time"

	shttp "github.com/skydive-project/skydive/graffiti/http"
	"github.com/skydive-project/skydive/graffiti/service"
	"github.com/skydive-project/skydive/graffiti/websocket"
)

// NewWSClientOpts creates WebSocket options object from the configuration
func NewWSClientOpts(authOpts *shttp.AuthenticationOpts) (*websocket.ClientOpts, error) {
	// override some of the options with config value
	tlsConfig, err := GetTLSClientConfig(true)
	if err != nil {
		return nil, err
	}

	return &websocket.ClientOpts{
		QueueSize:        GetInt("http.ws.queue_size"),
		WriteCompression: GetBool("http.ws.enable_write_compression"),
		TLSConfig:        tlsConfig,
		AuthOpts:         authOpts,
		Headers:          http.Header{},
	}, nil
}

// NewWSClient creates a Client based on the configuration
func NewWSClient(clientType service.Type, url *url.URL, opts websocket.ClientOpts) (*websocket.Client, error) {
	host := GetString("host_id")

	return websocket.NewClient(host, clientType, url, opts), nil
}

// NewWSServerOpts returns WebSocket server options
func NewWSServerOpts() websocket.ServerOpts {
	pingDelay := time.Duration(GetInt("http.ws.ping_delay")) * time.Second

	return websocket.ServerOpts{
		WriteCompression: GetBool("http.ws.enable_write_compression"),
		QueueSize:        GetInt("http.ws.queue_size"),
		PingDelay:        pingDelay,
		PongTimeout:      time.Duration(GetInt("http.ws.pong_timeout"))*time.Second + pingDelay,
	}
}

// NewWSServer creates a Server based on the configuration
func NewWSServer(server *shttp.Server, endpoint string, authBackend shttp.AuthenticationBackend) *websocket.Server {
	opts := NewWSServerOpts()
	opts.AuthBackend = authBackend
	return websocket.NewServer(server, endpoint, opts)
}
