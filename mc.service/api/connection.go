package api

import (
	"net/http"
	"net/url"
	"time"
)

type Connection interface {
	Request(endpoint *url.URL) (*http.Response, error)
}

type ClientHost struct {
	Client *http.Client
	Host   string
}

type Client struct {
	Connection Connection
	ApiKey     string
}

func (conn *ClientHost) Request(endpoint *url.URL) (*http.Response, error) {
	endpoint.Scheme = "https"
	endpoint.Host = conn.Host
	targetUrl := endpoint.String()
	return conn.Client.Get(targetUrl)
}

func ClientFactory(host string, apiKey string, timeout time.Duration) *Client {
	client := &http.Client{
		Timeout: timeout,
	}
	
	clientHost := &ClientHost{
		Client: client,
		Host:   host,
	}

	return &Client{
		Connection: clientHost,
		ApiKey:     apiKey,
	}
}