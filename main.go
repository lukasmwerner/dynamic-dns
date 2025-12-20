package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	httpTriger := make(chan int)
	if IsHttpAPIEnabled(l) {
		go startHttpServer(httpTriger, ":2025")
	}
	refresh := make(chan int)
	ticker := time.NewTicker(interval)

	ctx, cancel := context.WithCancel(context.Background())
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				refresh <- 1
			case <-httpTriger:
				refresh <- 1
			}
		}
	}()

	config := GetCloudflareConfigData(l)
	wg := startCloudflareConfig(ctx, l, config, refresh)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	cancel()
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
