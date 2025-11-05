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
	client *http.Client
	host   string
}

type Client struct {
	connection Connection
	apiKey     string
}

func (conn *ClientHost) Request(endpoint *url.URL) (*http.Response, error) {
	endpoint.Scheme = "https"
	endpoint.Host = conn.host
	targetUrl := endpoint.String()
	return conn.client.Get(targetUrl)
}

func ClientFactory(host string, apiKey string, timeout time.Duration) *Client {
	client := &http.Client{
		Timeout: timeout,
	}
	
	clientHost := &ClientHost{
		client: client,
		host:   host,
	}

	return &Client{
		connection: clientHost,
		apiKey:     apiKey,
	}
}