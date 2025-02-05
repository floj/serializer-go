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

const hnSearchURL = "https://hn.algolia.com/api/v1/search?tags=front_page&hitsPerPage=30"
const hnStoryURL = "https://hn.algolia.com/api/v1/items"

type HNScraper struct {
	httpc *http.Client
}

func NewScraper(httpc *http.Client) (*HNScraper, error) {
	return &HNScraper{
		httpc: httpc,
	}, nil
}

func (s *HNScraper) Name() string {
	return model.ScraperHN
}

func (s *HNScraper) FetchItem(ctx context.Context, refId string) (model.Story, bool, error) {
	uri := hnStoryURL + "/" + refId
	slog.Debug("fetching HN story", "url", uri)
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, uri, nil)
	if err != nil {
		return model.Story{}, false, err
	}
	resp, err := s.httpc.Do(req)
	if err != nil {
		return model.Story{}, false, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return model.Story{}, false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return model.Story{}, false, fmt.Errorf("request not successful, expected status 200, got %d", resp.StatusCode)
	}

	itm := Item{}
	err = json.NewDecoder(resp.Body).Decode(&itm)
	if err != nil {
		return model.Story{}, true, err
	}

	return model.Story{
		RefID:       refId,
		Url:         itm.URL,
		Title:       itm.Title,
		Score:       int64(itm.Points),
		NumComments: int64(itm.NumComments()),
	}, true, nil
}

func (s *HNScraper) FetchItems(ctx context.Context) ([]model.Story, error) {
	uri := hnSearchURL
	slog.Debug("fetching HN stories", "url", uri)
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, uri, nil)
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
				PublishedAt: h.CreatedAt,
				RefID:       strconv.Itoa(h.StoryID),
				Title:       h.Title,
				Type:        t,
				Score:       int64(h.Points),
				Scraper:     model.ScraperHN,
				NumComments: int64(h.NumComments),
			}
			stories = append(stories, story)
		}
	}

	return stories, nil
}
