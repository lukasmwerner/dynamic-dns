package porkbun

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Record struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"content"`
	TTL      string `json:"ttl"`
	Priority string `json:"prio"`
}

type dnsRecordsResponse struct {
	Status     string   `json:"status"`
	Cloudflare string   `json:"cloudflare"`
	Records    []Record `json:"records"`
}

func (c *client) GetDNSRecords(domain string) ([]Record, error) {
	req, err := http.NewRequest("GET", apiRoot+"/dns/retrieve/"+domain, nil)
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
	response := dnsRecordsResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response.Records, nil
}

func (c *client) SetDNSRecord(domain string, id string, value string) error {

	body, _ := json.Marshal(struct {
		Key     string `json:"apikey"`
		Secret  string `json:"secretapikey"`
		Type    string `json:"type"`
		Content string `json:"content"`
	}{
		Content: value,
		Key:     c.key,
		Secret:  c.secret,
		Type:    "A",
	})
	payload := bytes.NewBuffer(body)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/dns/edit/%s/%s", apiRoot, domain, id), payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("non 200 status code: " + resp.Status)
	}
	defer resp.Body.Close()

	return nil
}
