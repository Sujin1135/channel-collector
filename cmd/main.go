package main

import (
	"channel-collector/internal/channel"
	collector2 "channel-collector/internal/collector"
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	ch := make(chan *channel.Channel, 1)
	collector := collector2.NewCollector()

	go collector.Collect([]string{"@xiae3067", "@ezcd"}, ch, &wg)

	for data := range ch {
		fmt.Println(data)
	}
}
