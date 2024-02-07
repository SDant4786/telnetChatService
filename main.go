package main

import (
	"chatservice/config"
	"chatservice/http"
	"chatservice/telnet"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	//Load config
	cfg, err := config.LoadConfig("./config.env")
	if err != nil {
		log.Fatalf("Could not load config file. Err: %s", err)
	}
	//Set up logging to file
	f, err := os.OpenFile(cfg.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	// Create shut down channel and signal for clean closure
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	//Wait group to ensure telnet server is running before http server spins up
	wg := sync.WaitGroup{}
	wg.Add(1)
	//Spin up telnet server
	go telnet.InitTelnetServer(cfg, shutdown, &wg)
	wg.Wait()
	go http.InitHttpServer(cfg)
	<-shutdown
}
