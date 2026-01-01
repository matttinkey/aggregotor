package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

func (r *RSSFeed) clean() {
	r.Channel.Title = html.UnescapeString(r.Channel.Title)
	r.Channel.Description = html.UnescapeString(r.Channel.Description)
	for _, item := range r.Channel.Item {
		item.clean()
	}
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (r *RSSItem) clean() {
	r.Title = html.UnescapeString(r.Title)
	r.Description = html.UnescapeString(r.Description)
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}
	req.Header.Set("User-Agent", "gator")

	result, err := http.DefaultClient.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}

	if result.StatusCode > 399 {
		return &RSSFeed{}, fmt.Errorf("server error: %v", result.StatusCode)
	}

	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return &RSSFeed{}, err
	}

	var feed RSSFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return &RSSFeed{}, err
	}

	return &feed, nil
}
