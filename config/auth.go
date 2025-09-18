/*
 * Copyright (c) KylinSoft  Co., Ltd. 2025.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2.
 * See LICENSE file for more details.
 * Author: waneng <wanfeng@kylinos.cn>
 */
package config

import (
	"errors"
	"fmt"
	"os"

	auth "github.com/abbot/go-http-auth"
	shttp "github.com/skydive-project/skydive/graffiti/http"
)

// NewAuthenticationBackendByName creates a new auth backend based on the name
func NewAuthenticationBackendByName(name string) (backend shttp.AuthenticationBackend, err error) {
	typ := GetString("auth." + name + ".type")
	switch typ {
	case "basic":
		role := GetString("auth." + name + ".role")
		if role == "" {
			role = shttp.DefaultUserRole
		}

		var provider auth.SecretProvider
		if file := GetString("auth." + name + ".file"); file != "" {
			if _, err := os.Stat(file); err != nil {
				return nil, err
			}

			provider = auth.HtpasswdFileProvider(file)
		} else if users := GetStringMapString("auth." + name + ".users"); users != nil && len(users) > 0 {
			provider = shttp.NewHtpasswdMapProvider(users).SecretProvider()
		} else {
			return nil, errors.New("No htpassword provider set, you set either file or inline sections")
		}

		backend, err = shttp.NewBasicAuthenticationBackend(name, provider, role)
	case "keystone":
		authURL := GetString("auth." + name + ".auth_url")
		domain := GetString("auth." + name + ".domain_name")
		tenant := GetString("auth." + name + ".tenant_name")

		role := GetString("auth." + name + ".role")
		if role == "" {
			role = shttp.DefaultUserRole
		}

		backend, err = shttp.NewKeystoneBackend(name, authURL, tenant, domain, role)
	case "noauth":
		backend = shttp.NewNoAuthenticationBackend()
	default:
		err = fmt.Errorf("Authentication type unknown or backend not defined for: %s", name)
	}

	if err != nil {
		return nil, err
	}
	return backend, nil
}
