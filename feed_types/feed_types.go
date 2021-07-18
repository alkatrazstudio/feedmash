// SPDX-License-Identifier: AGPL-3.0-only
// 🄯 2021, Alexey Parfenov <zxed@alkatrazstudio.net>

package feed_types

import (
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"net/url"
	"os"
)

const (
	Unknown = iota
	Http
	Youtube
)

type FeedTypeFuncs struct {
	RealUrl                     func(feedUrl url.URL) string
	SourceFeedItemToOutFeedItem func(item *gofeed.Item) *feeds.Item
}

func Detect(feedUrl url.URL) (int, *FeedTypeFuncs) {
	if IsYoutube(feedUrl) {
		return Youtube, &youtubeSourceFuncs
	}

	if IsHttp(feedUrl) {
		return Http, &httpSourceFuncs
	}

	_, _ = fmt.Fprintln(os.Stderr, "Can't determine feed type: "+feedUrl.String())
	return Unknown, nil
}
