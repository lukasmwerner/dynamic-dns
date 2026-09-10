package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

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

	httpTriger := make(chan int)
	if IsHttpAPIEnabled(l) {
		go startHttpServer(httpTriger, ":2025")
	}
	broadcastRefresh := NewBroadcaster(0)
	defer broadcastRefresh.Close()
	ticker := time.NewTicker(interval)

	notify := make(chan Notification)
	ctx, cancel := context.WithCancel(context.Background())
	defer ticker.Stop()
	go func() {
		// Main Signal Loop
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// PERF: Cache current IPv4 and only send signal when differs
				ipv4Address, err := FetchConfigIPv4(l)
				if err != nil {
					log.Printf("unable to fetch ipv4 address in lua: %s\n", err.Error())
					continue
				}
				broadcastRefresh.Submit(ipv4Address)
			case <-httpTriger:
				ipv4Address, err := FetchConfigIPv4(l)
				if err != nil {
					log.Printf("unable to fetch ipv4 address in lua: %s\n", err.Error())
					continue
				}
				broadcastRefresh.Submit(ipv4Address)
			case notification := <-notify:
				Notify(l, notification)
			}
		}
	}()

	var wg sync.WaitGroup
	// BUG: when both Cloudflare and Porkbun are used, only one will get
	// channel notifications
	cloudflare_config := GetCloudflareConfigData(l)
	startCloudflareConfig(ctx, notify, cloudflare_config, &wg, broadcastRefresh)
	porkbun_config := GetPorkbunConfigData(l)
	startPorkbunConfig(ctx, notify, porkbun_config, &wg, broadcastRefresh)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs    // Wait for kill handler
	cancel()  // Kill main loop
	wg.Wait() // Wait for workers to clean up
}

func startHttpServer(sig chan int, addr string) {
	http.HandleFunc("GET /rpc/refresh_dns", func(w http.ResponseWriter, r *http.Request) {
		log.Println("got refresh signal via http. triggering refresh signal")
		sig <- 1
		w.WriteHeader(http.StatusProcessing)
	})
	http.ListenAndServe(addr, nil)
}
