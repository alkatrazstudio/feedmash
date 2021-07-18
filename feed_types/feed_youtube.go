// SPDX-License-Identifier: AGPL-3.0-only
// 🄯 2021, Alexey Parfenov <zxed@alkatrazstudio.net>

package feed_types

import (
	"bytes"
	"feedmash/util"
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"html/template"
	"net/url"
	"regexp"
	"strings"
)

func IsYoutube(feedUrl url.URL) bool {
	scheme := strings.ToLower(feedUrl.Scheme)
	if scheme != "https" || (feedUrl.Host != "youtube.com" && !strings.HasSuffix(feedUrl.Host, ".youtube.com")) {
		return false
	}

	return true
}

func YoutubeRealUrl(feedUrl url.URL) string {
	rx := regexp.MustCompile(`^/channel/([\w\-_]+)`)
	matches := rx.FindStringSubmatch(feedUrl.Path)
	if matches == nil {
		return ""
	}

	channelId := matches[1]

	realUrl := url.URL{
		Scheme: "https",
		Host:   "www.youtube.com",
		Path:   "feeds/videos.xml",
	}
	query := realUrl.Query()
	query.Set("channel_id", channelId)
	realUrl.RawQuery = query.Encode()

	realUrlStr := realUrl.String()
	return realUrlStr
}

func YoutubeSourceFeedItemToOutFeedItem(item *gofeed.Item) *feeds.Item {
	outItem := HttpSourceFeedItemToOutFeedItem(item)
	if outItem == nil {
		return nil
	}

	tplText := `
		<p><a href="{{.Href}}" target="_blank" rel="referrer"><img src="{{.ThumbnailUrl}}" /></a></p>
		<p>{{.Content}}</p>
	`
	tplText = strings.ReplaceAll(tplText, "\n", "")

	tpl, err := template.New("yt").Parse(tplText)
	if err != nil {
		util.LogWarn(err)
		return outItem
	}

	defer func() {
		if err := recover(); err != nil {
			util.LogWarn(fmt.Sprintf("%s: %s", item.Title, err))
		}
	}()

	thumbnailUrl := item.Extensions["media"]["group"][0].Children["thumbnail"][0].Attrs["url"]
	content := item.Extensions["media"]["group"][0].Children["description"][0].Value

	data := struct {
		Href         string
		ThumbnailUrl string
		Content      string
	}{
		Href:         outItem.Link.Href,
		ThumbnailUrl: thumbnailUrl,
		Content:      content,
	}

	var renderedBytes bytes.Buffer
	err = tpl.Execute(&renderedBytes, data)
	if err != nil {
		util.LogWarn(err)
		return outItem
	}

	outItem.Content = strings.TrimSpace(renderedBytes.String())
	outItem.Content = strings.ReplaceAll(outItem.Content, "\n", "<br/>")
	outItem.Description = outItem.Content
	return outItem
}

var youtubeSourceFuncs = FeedTypeFuncs{
	RealUrl:                     YoutubeRealUrl,
	SourceFeedItemToOutFeedItem: YoutubeSourceFeedItemToOutFeedItem,
}
