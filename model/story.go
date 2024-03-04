package model

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/dustin/go-humanize"
)

func subpath(p string, i int) string {
	parts := strings.SplitN(p, "/", i+2)
	if len(parts) <= i+1 {
		return p
	}
	return strings.Join(parts[0:i+1], "/")
}

func (s *Story) Domain() string {
	if s.Url == "" {
		return ""
	}

	u, err := url.Parse(s.Url)
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	switch host {
	case "github.com":
		return host + subpath(u.Path, 2)
	case "gitlab.com":
		return host + subpath(u.Path, 2)
	case "twitter.com":
		return host + subpath(u.Path, 1)
	}
	return host
}

func (s *Story) TimeAgo() string {
	return humanize.Time(s.CreatedAt)
}

func (s *Story) LinkURL() string {
	u := "#"

	switch s.Scraper {
	case ScraperHN:
		switch s.Type {
		case TypeHNAskHN:
			u = s.CommentsURL()
		default:
			u = s.Url
		}
	}

	if u == "#" {
		return u
	}

	pu, err := url.Parse(u)
	if err != nil {
		return u
	}

	for _, t := range urlTransformer {
		if t.Matches(pu) {
			return t.Transform(pu).String()
		}
	}

	return u
}

func (s *Story) SearchURL() string {
	switch s.Scraper {
	case ScraperHN:
		return "https://hn.algolia.com/?dateRange=pastYear&type=story&query=" + url.QueryEscape(s.Title)
	default:
		return "#"
	}
}

func (s *Story) CommentsURL() string {
	switch s.Scraper {
	case ScraperHN:
		return "https://news.ycombinator.com/item?id=" + url.QueryEscape(s.RefID)
	default:
		return "#"
	}
}

func (s *Story) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

func Deserialize(b []byte) (Story, error) {
	s := Story{}
	err := json.Unmarshal(b, &s)
	return s, err
}
