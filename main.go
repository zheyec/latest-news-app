package main

import (
	"context"
	"flag"
	"fmt"
	"lxm-news/service"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	var port int
	flag.IntVar(&port, "p", 8000, "端口号")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		service.Run(ctx, port)
	}()

	go func() {
		sig := <-sigs
		fmt.Println(sig)
		cancel()
		done <- true
	}()

	var bc = new(service.BackgroundCrawler)
	go bc.Start()

	service.RemoveAllTemp(service.RawNewsPath)
	service.RemoveAllTemp(service.TimgPath)
	service.RemoveAllTemp(service.CardPath)

	<-done
	fmt.Println("[IAM]=>stop service.")
}
