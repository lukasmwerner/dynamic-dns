package main

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zones"
	"github.com/joho/godotenv"
)

type DNSRecord struct {
	Zone    string
	ID      string
	Name    string
	Content string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("unable to load config")
	}
	l := LoadLua("config.lua")
	ip, err := FetchConfigIPv4(l)
	if err != nil {
		return
	}
	fmt.Println("current ip:", ip)

	interval := GetConfigInterval(l)
	fmt.Println("checking every:", interval)

	config := GetCloudflareConfigData(l)

	records := []DNSRecord{}
	var client *cloudflare.Client

	for token, domains := range config {
		client = cloudflare.NewClient(option.WithAPIToken(token))
		z, err := client.Zones.List(context.Background(), zones.ZoneListParams{})
		if err != nil {
			log.Println("encountered error accessing zones:", err.Error())
			continue
		}
		for _, zone := range z.Result {
			fmt.Printf("%s\n", zone.Name)
			r, err := client.DNS.Records.List(context.Background(), dns.RecordListParams{
				ZoneID: cloudflare.F(zone.ID),
				Type:   cloudflare.F(dns.RecordListParamsTypeA),
			})
			if err != nil {
				log.Println("encountered error accessing dns record:", err.Error())
				continue
			}
			for _, record := range r.Result {
				fmt.Printf("\t%s %s %s\n", record.Type, record.Name, record.Content)

				if slices.Contains(domains, record.Name) {
					records = append(records, DNSRecord{
						Zone:    zone.ID,
						ID:      record.ID,
						Name:    record.Name,
						Content: record.Content,
					})
				}
			}
		}
	}

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			log.Println("interval starting")
			ipv4Address, err := FetchConfigIPv4(l)
			if err != nil {
				log.Printf("unable to fetch ipv4 address in lua: %s\n", err.Error())
				continue
			}
			for i, record := range records {
				if record.Content == ipv4Address {
					continue
				}
				log.Printf("changing %s, ipv4 different (%s!=%s) changing to %s\n", record.Name, record.Content, ipv4Address, ipv4Address)
				res, err := client.DNS.Records.Edit(context.Background(), record.ID, dns.RecordEditParams{
					ZoneID: cloudflare.F(record.Zone),
					Body: dns.ARecordParam{
						Name:    cloudflare.F(record.Name),
						Type:    cloudflare.F(dns.ARecordTypeA),
						Content: cloudflare.F(ipv4Address),
					},
				})
				if err != nil {
					log.Printf("error changing record for %s: %s\n", record.Name, err.Error())
					continue
				}
				records[i].Content = res.Content
			}
		}
	}
}
