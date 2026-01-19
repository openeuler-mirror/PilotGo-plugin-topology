package config

import (
	"bytes"
	"testing"

	capturer "github.com/zenizh/go-capturer"
)

func TestRelocation(t *testing.T) {
	cfg.SetConfigType("yaml")

	relocationMap = map[string][]string{
		"openstack.auth_url": {
			"auth.keystone.auth_url",
			"agent.topology.neutron.auth_url",
		},
		"openstack.tenant_name": {
			"auth.keystone.tenant_name",
			"agent.topology.neutron.tenant_name",
		},
	}

	var yamlv1 = []byte(`
auth:
  keystone:
    auth_url: "http://test.test"
    tenant_name: "test"
`)

	cfg.ReadConfig(bytes.NewBuffer(yamlv1))
	out := capturer.CaptureStderr(func() {
		value := Get("openstack.auth_url")
		if value != "http://test.test" {
			t.Fatal("Relocation failed")
		}
	})
	if len(out) == 0 {
		t.Fatal("A warning should have been raised")
	}

	value := Get("openstack.tenant_name")
	if value != "test" {
		t.Fatal("Relocation with default failed")
	}

	var yamlv2 = []byte(`
openstack:
  auth_url: "http://test.test"
`)

	cfg.ReadConfig(bytes.NewBuffer(yamlv2))
	out = capturer.CaptureStderr(func() {
		value := Get("openstack.auth_url")
		if value != "http://test.test" {
			t.Fatal("Relocation failed")
		}
	})
	if len(out) != 0 {
		t.Fatal("A warning shouldn't have been raised")
	}

	value = Get("openstack.tenant_name")
	if value != "admin" {
		t.Fatal("Relocation with default failed")
	}
}
