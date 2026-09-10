package porkbun

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type getDomainsResponse struct {
	Status  string   `json:"status"`
	Count   int      `json:"count"`
	Domains []Domain `json:"domains"`
}

type Domain struct {
	Domain string `json:"domain"`
	Status string `json:"status"`
	TLD    string `json:"tld"`
}

func (c *client) GetDomains() ([]Domain, error) {
	req, err := http.NewRequest("GET", apiRoot+"/domain/listAll", nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("non 200 status code: " + resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	domainResponse := getDomainsResponse{}
	err = json.Unmarshal(body, &domainResponse)
	if err != nil {
		return nil, err
	}

	return domainResponse.Domains, nil
}
