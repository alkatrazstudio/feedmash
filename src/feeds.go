// SPDX-License-Identifier: AGPL-3.0-only
// 🄯 2021, Alexey Parfenov <zxed@alkatrazstudio.net>

package src

import (
	"encoding/xml"
	"feedmash/feed_types"
	"feedmash/util"
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FeedSource struct {
	url      string
	urlObj   url.URL
	feedType int
	funcs    *feed_types.FeedTypeFuncs
	realUrl  string
	stop     chan bool
	stopped  chan bool
}

type FeedChanItem struct {
	source FeedSource
	feed   gofeed.Feed
}

func newFeedSource(feedUrl string) *FeedSource {
	urlObj, err := url.Parse(feedUrl)
	if err != nil {
		util.LogWarn(err)
		return nil
	}

	feedType, funcs := feed_types.Detect(*urlObj)
	if feedType == feed_types.Unknown {
		return nil
	}

	source := FeedSource{
		url:      feedUrl,
		urlObj:   *urlObj,
		feedType: feedType,
		funcs:    funcs,
		realUrl:  "",
		stop:     make(chan bool),
		stopped:  make(chan bool),
	}

	return &source
}

func loadSourceFeed(feedSource FeedSource, cfg Config) *gofeed.Feed {
	fp := gofeed.NewParser()
	fp.UserAgent = cfg.userAgent

	if feedSource.realUrl == "" {
		feedSource.realUrl = feedSource.funcs.RealUrl(feedSource.urlObj)
		if feedSource.realUrl == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Cannot get real url: "+feedSource.url)
			return nil
		}
	}

	feed, err := fp.ParseURL(feedSource.realUrl)
	if err != nil {
		util.LogWarn(fmt.Sprintf("%s (%s) %s", feedSource.realUrl, feedSource.url, err))
		return nil
	}
	return feed
}

func appendFeedItem(curOutItems []*feeds.Item, item *feeds.Item) []*feeds.Item {
	for _, outFeedItem := range curOutItems {
		if outFeedItem.Id == item.Id {
			return curOutItems
		}
	}

	items := append(curOutItems, item)
	sort.Slice(items, func(a, b int) bool {
		return items[a].Created.After(items[b].Created)
	})

	return items
}

func randDurationInRange(minMins int, maxMins int) time.Duration {
	mins := rand.Int63n(int64(maxMins-minMins)) + int64(minMins)
	interval := time.Duration(mins)
	return interval
}

func watchFeed(feedSource FeedSource, cfg Config, sourceFeedsChan chan *FeedChanItem, initialPause time.Duration) {
	timer := time.NewTimer(initialPause)

	for {
		select {
		case <-feedSource.stop:
			feedSource.stopped <- true
			return

		case <-timer.C:
			feed := loadSourceFeed(feedSource, cfg)
			if feed == nil {
				break
			}
			sourceFeedsChan <- &FeedChanItem{
				source: feedSource,
				feed:   *feed,
			}
			newInterval := randDurationInRange(cfg.minIntervalMins, cfg.maxIntervalMins) * time.Minute
			timer = time.NewTimer(newInterval)
		}
	}
}

func startWatchingFeeds(feedSources []FeedSource, cfg Config, sourceFeedsChan chan *FeedChanItem) {
	for feedIndex, feedSource := range feedSources {
		var initialPause = time.Duration(cfg.initialPauseSecs*feedIndex) * time.Second
		go watchFeed(feedSource, cfg, sourceFeedsChan, initialPause)
	}
}

func stopWatchingFeeds(feedSources []FeedSource) {
	for _, feedSource := range feedSources {
		feedSource.stop <- true
	}
	for _, feedSource := range feedSources {
		<-feedSource.stopped
	}
}

func loadSources(urls []string) []FeedSource {
	var feedSources []FeedSource
	for _, feedUrl := range urls {
		feedSource := newFeedSource(feedUrl)
		if feedSource == nil {
			continue
		}
		feedSources = append(feedSources, *feedSource)
	}
	return feedSources
}

func saveToFile(filename string, s string) {
	tmpFilename := filename + ".tmp"

	outFeedDir := filepath.Dir(tmpFilename)
	err := os.MkdirAll(outFeedDir, 0750)
	if err != nil {
		util.LogWarn(err)
		return
	}

	f, err := os.Create(tmpFilename)
	if err != nil {
		util.LogWarn(err)
		return
	}

	isClosed := false
	defer func() {
		if isClosed {
			return
		}
		if err := f.Close(); err != nil {
			util.LogWarn(err)
		}
	}()

	_, err = f.WriteString(s)
	if err != nil {
		util.LogWarn(err)
		return
	}

	err = f.Close()
	if err != nil {
		util.LogWarn(err)
		return
	}
	isClosed = true

	err = os.Rename(tmpFilename, filename)
	if err != nil {
		util.LogWarn(err)
		return
	}
}

func mergeOutFeedItems(
	oldItems []*feeds.Item,
	newItems []*gofeed.Item,
	maxOutItems int,
	sourceFeedItemToOutFeedItem func(item *gofeed.Item) *feeds.Item,
) []*feeds.Item {
	resultItems := oldItems

	for _, item := range newItems {
		outItem := sourceFeedItemToOutFeedItem(item)
		if outItem != nil {
			resultItems = appendFeedItem(resultItems, outItem)
		}
	}

	nItems := len(resultItems)
	if nItems > maxOutItems {
		resultItems = resultItems[0:maxOutItems]
	}

	return resultItems
}

func feedToStr(feed *feeds.Feed) string {
	xmlInternal := feeds.Atom{Feed: feed}
	xmlFeed := xmlInternal.AtomFeed()
	xmlFeed.Id = feed.Id
	data, err := xml.Marshal(xmlFeed)
	if err != nil {
		util.LogWarn(err)
		return ""
	}
	xmlStr := strings.TrimSpace(xml.Header) + string(data)
	return xmlStr
}

func startSourceFeedsReceiver(
	cfg Config,
	feedsChan chan *FeedChanItem,
	sourceFeedsReceiverStopped chan bool,
	outXmlChan chan string,
) {
	outFeedData := loadOutFeed(cfg)

	outFeed := &feeds.Feed{
		Id:    cfg.outFeedId,
		Title: cfg.outFeedTitle,
		Link: &feeds.Link{
			Href: cfg.outFeedSelfLink,
			Rel:  "self",
		},
		Items: []*feeds.Item{},
	}

	changed := false
	if outFeedData != nil {
		outFeed.Items = mergeOutFeedItems(
			outFeed.Items,
			outFeedData.Items,
			cfg.maxOutItems,
			feed_types.HttpSourceFeedItemToOutFeedItem,
		)

		if outFeedData.UpdatedParsed != nil && len(outFeed.Items) == len(outFeedData.Items) {
			outFeed.Updated = *outFeedData.UpdatedParsed
			changed = false
		}
	}

	if changed {
		outFeed.Updated = time.Now()
	}

	newOutXml := feedToStr(outFeed)

	if newOutXml == "" {
		util.LogWarn("Can't generate out XML.")
		sourceFeedsReceiverStopped <- true
		return
	}

	outXmlChan <- newOutXml

	if changed {
		saveToFile(cfg.outFeedFilename, newOutXml)
	}

	for {
		chanItem := <-feedsChan
		if chanItem == nil {
			break
		}

		oldIds := []string{}
		for _, item := range outFeed.Items {
			oldIds = append(oldIds, item.Id)
		}

		outFeed.Items = mergeOutFeedItems(
			outFeed.Items,
			chanItem.feed.Items,
			cfg.maxOutItems,
			chanItem.source.funcs.SourceFeedItemToOutFeedItem,
		)

		changed = false
		for i, item := range outFeed.Items {
			if len(oldIds) <= i {
				changed = true
				break
			}
			if item.Id != oldIds[i] {
				changed = true
				break
			}
		}
		if !changed {
			continue
		}

		outFeed.Updated = time.Now()

		newOutXml = feedToStr(outFeed)
		if newOutXml == "" {
			continue
		}

		outXmlChan <- newOutXml

		saveToFile(cfg.outFeedFilename, newOutXml)
	}

	sourceFeedsReceiverStopped <- true
}

func loadOutFeed(cfg Config) *gofeed.Feed {
	if _, err := os.Stat(cfg.outFeedFilename); err != nil {
		util.LogWarn(err)
		return nil
	}

	file, err := os.Open(cfg.outFeedFilename)
	defer func() {
		if err := file.Close(); err != nil {
			util.LogWarn(err)
		}
	}()

	if err != nil {
		util.LogWarn(err)
		return nil
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(file)
	if err != nil {
		util.LogWarn(err)
		return nil
	}

	return feed
}
