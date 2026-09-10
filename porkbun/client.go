package porkbun

import "net/http"

var apiRoot = "https://api.porkbun.com/api/json/v3"

type client struct {
	http   *http.Client
	key    string
	secret string
}

func NewClient(key string, secret string) (*client, error) {
	client := &client{
		http:   http.DefaultClient,
		key:    key,
		secret: secret,
	}

	return client, nil
}

func (c *client) setHeaders(r *http.Request) {
	r.Header.Set("X-API-Key", c.key)
	r.Header.Set("X-Secret-API-Key", c.secret)
}
