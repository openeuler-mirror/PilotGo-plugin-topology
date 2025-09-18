/*
 * Copyright (c) KylinSoft  Co., Ltd. 2025.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2.
 * See LICENSE file for more details.
 * Author: waneng <wanfeng@kylinos.cn>
 */
package config

import (
	"crypto/tls"

	gtls "github.com/skydive-project/skydive/graffiti/tls"
)

// GetTLSClientConfig returns TLS config to be used by client
func GetTLSClientConfig(setupRootCA bool) (*tls.Config, error) {
	certPEM := GetString("tls.client_cert")
	keyPEM := GetString("tls.client_key")
	var tlsConfig *tls.Config
	if certPEM != "" && keyPEM != "" {
		var err error
		tlsConfig, err = gtls.SetupTLSClientConfig(certPEM, keyPEM)
		if err != nil {
			return nil, err
		}
		if setupRootCA {
			rootCaPEM := GetString("tls.ca_cert")
			tlsConfig.RootCAs, err = gtls.SetupTLSLoadCA(rootCaPEM)
			if err != nil {
				return nil, err
			}
		}
	}
	return tlsConfig, nil
}

// GetTLSServerConfig returns TLS config to be used by server
func GetTLSServerConfig(setupRootCA bool) (*tls.Config, error) {
	certPEM := GetString("tls.server_cert")
	keyPEM := GetString("tls.server_key")

	tlsConfig, err := gtls.SetupTLSServerConfig(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	if GetBool("tls.require_client_cert") {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if setupRootCA {
		rootCaPEM := GetString("tls.ca_cert")
		tlsConfig.ClientCAs, err = gtls.SetupTLSLoadCA(rootCaPEM)
		if err != nil {
			return nil, err
		}
	}
	return tlsConfig, nil
}
