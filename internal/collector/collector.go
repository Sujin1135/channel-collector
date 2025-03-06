package collector

import (
	"channel-collector/internal/channel"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"sync"
	"time"
)

const (
	extractScript = `
		function convertAbbreviatedNumber(str) {
		  let multiplier = 1;
		  let numStr = str.trim();

		  if (numStr.endsWith("k") || numStr.endsWith("K")) {
			multiplier = 1000;
			numStr = numStr.slice(0, -1);
		  } else if (numStr.endsWith("M")) {
			multiplier = 1000000;
			numStr = numStr.slice(0, -1);
		  } else if (numStr.endsWith("B")) {
			multiplier = 1000000000;
			numStr = numStr.slice(0, -1);
		  }

		  return parseInt(numStr) * multiplier;
		}
		(function() {
			const date = new Date(Array.from(document.querySelectorAll("#additional-info-container tbody tr td")).filter((el) => el.innerText.includes("Joined "))[0].innerText.replace("Joined ", ""));
			const headerMetadata = ytInitialData.header.pageHeaderRenderer.content.pageHeaderViewModel.metadata.contentMetadataViewModel.metadataRows[1].metadataParts.map((v) => v.text.content);
			const metadata = ytInitialData.metadata.channelMetadataRenderer;
			return {
				externalId: metadata.externalId,
				title: metadata.title,
				description: metadata.description,
				isFamilySafe: metadata.isFamilySafe,
				keywords: metadata.keywords,
				thumbnails: metadata.avatar.thumbnails.map((v) => v.url),
				links: Array.from(document.querySelectorAll("#links-section #link-list-container yt-channel-external-link-view-model div span"), el => el.innerText),
				viewCount: Number(Array.from(document.querySelectorAll("#additional-info-container tbody tr td")).filter((el) => el.innerText.includes(" views"))[0].innerText.replace(" views", "").replaceAll(",", "")),
				totalSubscriber: convertAbbreviatedNumber(headerMetadata[0].replace(" subscribers", "")),
				totalVideo: convertAbbreviatedNumber(headerMetadata[1].replace(" videos", "")),
				joined: {
					year: date.getFullYear(),
					month: date.getMonth() + 1,
					date: date.getDate(),
				},
			};
		})();
	`
)

type Collector struct {
	rateLimiter <-chan time.Time
}

func NewCollector() *Collector {
	return &Collector{
		rateLimiter: time.Tick(1 * time.Second),
	}
}

func (c *Collector) Collect(youtubeHandles []string, ch chan<- *channel.Channel, errCh chan<- error, wg *sync.WaitGroup) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	defer close(ch)
	defer close(errCh)
	defer wg.Done()

	for _, youtubeHandle := range youtubeHandles {
		<-c.rateLimiter

		log.Printf("start to access website by handle: %s\n", youtubeHandle)

		var response *channel.Channel
		runErr := chromedp.Run(ctx,
			chromedp.Navigate(c.genChannelURL(youtubeHandle)),
			chromedp.WaitVisible(".truncated-text-wiz__absolute-button", chromedp.NodeVisible),
			chromedp.Click(".truncated-text-wiz__absolute-button", chromedp.NodeVisible),
			chromedp.WaitVisible("#additional-info-container", chromedp.ByQuery),
			chromedp.Evaluate(extractScript, &response),
		)
		if runErr != nil {
			errCh <- runErr
			continue
		}

		ch <- response

		log.Printf("end to access website by handle: %s\n", youtubeHandle)
	}
}

func (c *Collector) genChannelURL(youtubeHandle string) string {
	return fmt.Sprintf("https://www.youtube.com/%s", youtubeHandle)
}
