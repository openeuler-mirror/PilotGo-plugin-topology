package http

import (
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// RestClient describes a REST API client with a URL and authentication information
type RestClient struct {
	authOpts *AuthenticationOpts
	client   *http.Client
	url      *url.URL
}

// CrudClient describes a REST API client to issue CRUD commands
type CrudClient struct {
	*RestClient
}

// CreateOptions describes the options available when creating a resource
type CreateOptions struct {
	TTL time.Duration
}

func readBody(resp *http.Response) string {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(data)
}

func getHTTPClient(tlsConfig *tls.Config) *http.Client {
	client := &http.Client{}
	if tlsConfig != nil {
		tr := &http.Transport{TLSClientConfig: tlsConfig}
		client = &http.Client{Transport: tr}
	}
	return client
}

// MakeURL creates an URL for the specified protocol, address, port and path,
// whether TLS is required or not
func MakeURL(protocol string, addr string, port int, path string, useTLS bool) (*url.URL, error) {
	u, err := url.Parse(fmt.Sprintf("%s://%s:%d%s", protocol, addr, port, path))
	if err != nil {
		return nil, err
	}

	if (protocol == "http" || protocol == "ws") && useTLS {
		u.Scheme += "s"
	}

	return u, nil
}

// NewRestClient returns a new REST API client. It takes a URL
// to the HTTP point, authentication information and TLS configuration
func NewRestClient(url *url.URL, authOpts *AuthenticationOpts, tlsConfig *tls.Config) *RestClient {
	return &RestClient{
		client:   getHTTPClient(tlsConfig),
		url:      url,
		authOpts: authOpts,
	}
}

// Request issues a request to the API
func (c *RestClient) Request(method, path string, body io.Reader, header http.Header) (*http.Response, error) {
	url := c.url.ResolveReference(&url.URL{Path: path})
	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, err
	}

	if header != nil {
		req.Header = header
	}

	if c.authOpts != nil {
		SetAuthHeaders(&req.Header, c.authOpts)
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Add("Accept-Encoding", "gzip")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.ContentLength != 0 {
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			resp.Body, err = gzip.NewReader(resp.Body)
			resp.Uncompressed = true
			resp.ContentLength = -1
			if err != nil {
				return nil, err
			}
		}
	}

	return resp, nil
}
