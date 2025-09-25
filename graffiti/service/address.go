package service

import (
	"fmt"
	"net"
	"strings"
)

// Address describes the service listening address and port
type Address struct {
	Addr string
	Port int
}

func (sa Address) String() string {
	return fmt.Sprintf("%s:%d", sa.Addr, sa.Port)
}

// AddressFromString returns a service address from a string, could be IPv4 or IPv6
func AddressFromString(addressPort string) (Address, error) {
	/* Backward compatibility for old format like : listen = 1234 */
	if !strings.ContainsAny(addressPort, ".:") {
		addressPort = "localhost:" + addressPort
	} else if strings.HasPrefix(addressPort, ":") {
		addressPort = "localhost" + addressPort
	}

	host, port, err := net.SplitHostPort(addressPort)
	if err != nil {
		return Address{}, err
	}

	portNum, err := net.LookupPort("", port)
	if err != nil {
		return Address{}, err
	}

	return Address{
		Addr: host,
		Port: portNum,
	}, nil
}

// normalizeIPForURL returns a string normalized that can be used in URL. Brackets
// will be used for IPV6 addresses.
func normalizeIPForURL(ip net.IP) string {
	if ip.To4() == nil {
		return "[" + ip.String() + "]"
	}
	return ip.String()
}
