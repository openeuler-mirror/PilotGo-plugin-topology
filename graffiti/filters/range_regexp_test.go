package filters

import (
	"fmt"
	"regexp"
	"testing"
)

func TestIPV4Range24(t *testing.T) {
	expr, err := IPV4CIDRToRegex("192.168.0.0/24")
	if err != nil {
		t.Error(err)
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i != 254; i++ {
		ip := fmt.Sprintf("192.168.0.%d/24", i)

		if !re.MatchString(ip) {
			t.Errorf("%s not matching the rexp %s", ip, expr)
		}
	}

	ip := "192.168.1.0/24"
	if re.MatchString(ip) {
		t.Errorf("%s matches the rexp %s", ip, expr)
	}

	ip = "192.168.1.34/24"
	if re.MatchString(ip) {
		t.Errorf("%s matches the rexp %s", ip, expr)
	}
}

func TestIPV4Range16(t *testing.T) {
	expr, err := IPV4CIDRToRegex("192.168.0.0/16")
	if err != nil {
		t.Error(err)
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i != 254; i++ {
		for j := 0; j != 254; j++ {
			ip := fmt.Sprintf("192.168.0.%d/24", i)

			if !re.MatchString(ip) {
				t.Errorf("%s not matching the rexp %s", ip, expr)
			}
		}
	}

	ip := "192.169.3.34/24"
	if re.MatchString(ip) {
		t.Errorf("%s matches the rexp %s", ip, expr)
	}
}
