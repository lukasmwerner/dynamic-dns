package main

import (
	"context"
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/lukasmwerner/dynamic-dns/porkbun"
)

type KeyPair struct {
	Key    string
	Secret string
}

type PorkbunRecord struct {
	ID      string
	Domain  string
	Content string
}

func startPorkbunConfig(ctx context.Context, notify chan Notification, config map[KeyPair][]string, wg *sync.WaitGroup, refresh Broadcaster) {
	for keypair, domains := range config {
		// PERF: convert domains to a set to improve lookup for matches
		cli, err := porkbun.NewClient(keypair.Key, keypair.Secret)
		if err != nil {
			log.Println("encountered error creating porkbun client:", err.Error())
			continue
		}
		tracking_records := []PorkbunRecord{}
		pork_domains, err := cli.GetDomains()
		if err != nil {
			log.Println("encountered error accessing domains:", err.Error())
			continue
		}
		for _, domain := range pork_domains {
			fmt.Printf("%s\n", domain.Domain)
			records, err := cli.GetDNSRecords(domain.Domain)
			if err != nil {
				log.Println("encountered error accessing dns records:", err.Error())
				continue
			}
			for _, record := range records {
				if record.Type != "A" {
					continue
				}
				if slices.Contains(domains, record.Name) {
					fmt.Printf("\t%s %s %s\n", record.Type, record.Name, record.Value)
					tracking_records = append(tracking_records, PorkbunRecord{
						ID:      record.ID,
						Domain:  record.Name,
						Content: record.Value,
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
					log.Printf("interval starting for [%s]\n", keypair.Key[:8])
					for i, record := range tracking_records {
						if record.Content == ipv4Address {
							continue
						}
						log.Printf("changing %s, ipv4 different (%s!=%s) changing to %s\n", record.Domain, record.Content, ipv4Address, ipv4Address)
						err = cli.SetDNSRecord(record.Domain, record.ID, ipv4Address)
						if err != nil {
							log.Printf("error changing record for %s: %s\n", record.Domain, err.Error())
							continue
						}
						notify <- Notification{Domain: record.Domain, OldIP: record.Content, NewIP: ipv4Address}
						tracking_records[i].Content = ipv4Address
					}
				}
			}
		}(ctx, wg)
	}

}
