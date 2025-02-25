package main

import (
	"channel-collector/internal/channel"
	collector2 "channel-collector/internal/collector"
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	var dataWg sync.WaitGroup
	dataWg.Add(2)

	youtubeHandles := []string{"@xiae3067", "@ezcd"}
	ch := make(chan *channel.Channel, len(youtubeHandles))
	errCh := make(chan error, len(youtubeHandles))
	collector := collector2.NewCollector()

	go collector.Collect(youtubeHandles, ch, errCh, &wg)

	go func() {
		defer dataWg.Done()
		for err := range errCh {
			fmt.Println(errors.Wrap(err, "occurred an error when collect youtube channels"))
		}
	}()

	go func() {
		defer dataWg.Done()
		for data := range ch {
			fmt.Println("data as below:")
			fmt.Println(data)
		}
	}()

	wg.Wait()
	dataWg.Wait()
}
