// SPDX-License-Identifier: AGPL-3.0-only
// 🄯 2021, Alexey Parfenov <zxed@alkatrazstudio.net>

package feed_types

import (
	"feedmash/util"
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"net/url"
	"strings"
	"time"
)

func IsHttp(feedUrl url.URL) bool {
	scheme := strings.ToLower(feedUrl.Scheme)
	if scheme == "http" || scheme == "https" {
		return true
	}

	return false
}

func HttpRealUrl(feedUrl url.URL) string {
	return feedUrl.String()
}

func HttpSourceFeedItemToOutFeedItem(item *gofeed.Item) *feeds.Item {
	if item.Link == "" {
		util.LogWarn(fmt.Sprintf("Item \"%s\" has no link", item.Title))
		return nil
	}

	guid := item.GUID
	if guid == "" {
		guid = item.Link
		if item.PublishedParsed != nil {
			guid = item.PublishedParsed.Format("20060102-150405") + "-" + item.Link
		}
	}

	content := item.Content
	if content == "" {
		content = item.Description
	}
	description := item.Description
	if description == "" {
		description = content
	}

	published := item.PublishedParsed
	if published == nil {
		now := time.Now()
		published = &now
	}

	var author *feeds.Author = nil
	if len(item.Authors) > 0 {
		author = &feeds.Author{
			Name:  item.Authors[0].Name,
			Email: item.Authors[0].Email,
		}
	}

	outItem := feeds.Item{
		Id:    guid,
		Title: item.Title,
		Link: &feeds.Link{
			Href: item.Link,
		},
		Description: description,
		Author:      author,
		Created:     published.Local(),
		Content:     content,
	}
	return &outItem
}

var httpSourceFuncs = FeedTypeFuncs{
	RealUrl:                     HttpRealUrl,
	SourceFeedItemToOutFeedItem: HttpSourceFeedItemToOutFeedItem,
}
