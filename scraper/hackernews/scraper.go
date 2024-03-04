package hackernews

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/floj/serializer-go/model"
)

const hnURL = "https://hn.algolia.com/api/v1/search?tags=front_page&hitsPerPage=30"

type HNScraper struct {
	httpc *http.Client
}

func NewScraper(httpc *http.Client) (*HNScraper, error) {
	return &HNScraper{
		httpc: httpc,
	}, nil
}

func (s *HNScraper) FetchItems(ctx context.Context) ([]model.Story, error) {
	slog.Info("fetching HN stories", "url", hnURL)
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, hnURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request not successful, expected status 200, got %d", resp.StatusCode)
	}

	searchResult := SearchResult{}
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		return nil, err
	}

	hits := searchResult.Hits

	sort.Slice(hits, func(i, j int) bool {
		return hits[i].StoryID < hits[j].StoryID
	})

	stories := []model.Story{}
	now := time.Now()

	for _, h := range hits {
		t := h.GetType()
		switch t {
		case model.TypeHNJob:
			continue
		case model.TypeUnknown:
			slog.Warn("unknown type", "type", t, "scraper", "hn", "id", h.ObjectID)
			continue
		default:
			story := model.Story{
				By:          h.Author,
				Url:         h.URL,
				CreatedAt:   h.CreatedAt,
				ScrapedAt:   now,
				RefID:       strconv.Itoa(h.StoryID),
				Title:       h.Title,
				Type:        t,
				Score:       int32(h.Points),
				Scraper:     model.ScraperHN,
				NumComments: int32(h.NumComments),
			}
			stories = append(stories, story)
		}
	}

	return stories, nil
}
