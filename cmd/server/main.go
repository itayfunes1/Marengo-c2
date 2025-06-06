package main

import (
	"log"
	"os"
	"os/signal"
	"server-tcp-go/src/server"
	"server-tcp-go/src/work"
	"syscall"
)

func main() {
	server.Init()
	work.Init()
	go server.PrintMenu()
	stopOrKillServer()
}

func stopOrKillServer() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	sig := <-signals
	log.Println("\nReceive Signal from OS - Release resource")
	work.Stop()
	log.Println(sig)
	os.Exit(0)
}
