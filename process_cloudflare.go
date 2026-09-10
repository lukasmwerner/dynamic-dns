package main

import (
	"context"
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zones"
)

type CloudflareRecord struct {
	Zone    string
	ID      string
	Name    string
	Content string
}

func startCloudflareConfig(ctx context.Context, notify chan Notification, config map[string][]string, wg *sync.WaitGroup, refresh Broadcaster) {
	for token, domains := range config {
		records := []CloudflareRecord{}
		client := cloudflare.NewClient(option.WithAPIToken(token))
		println("new client")

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
					records = append(records, CloudflareRecord{
						Zone:    zone.ID,
						ID:      record.ID,
						Name:    record.Name,
						Content: record.Content,
					})
				}
			}
		}

		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup) {
			signal := make(chan any)
			refresh.Register(signal)
			defer refresh.Unregister(signal)

			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-signal:
					if !ok {
						return
					}
					ipv4Address := msg.(string)
					log.Printf("interval starting for [%s]\n", token[:8])
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
						notify <- Notification{Domain: record.Name, OldIP: record.Content, NewIP: ipv4Address}
						records[i].Content = res.Content
					}
				}
			}
		}(ctx, wg)
	}

}
