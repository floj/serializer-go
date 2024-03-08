package job

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/floj/serializer-go/config"
	"github.com/floj/serializer-go/model"
	"github.com/floj/serializer-go/scraper"
)

func Start(db *sql.DB, conf config.Config, scrapers ...scraper.Scraper) (func(func(Result, error) error) error, func()) {
	// restrict interval to be at max every minute
	interval := max(conf.ScrapeInterval, time.Minute)
	if !conf.ScrapeEnabled() {
		interval = time.Hour * 24
	}

	ticker := time.NewTicker(interval)
	quit := make(chan struct{})
	mu := &sync.Mutex{}

	go func() {
		for {
			select {
			case <-ticker.C:
				if conf.ScrapeEnabled() {
					runScrape(db, mu, conf, scrapers...)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return func(f func(Result, error) error) error {
			return f(runScrape(db, mu, conf, scrapers...))
		}, func() {
			quit <- struct{}{}
		}
}

type Result struct {
	New     int
	Updated int
	Errors  int
	err     []error
}

func (r *Result) Add(other Result) {
	r.New += other.New
	r.Updated += other.Updated
	r.err = append(r.err, other.err...)
	r.Errors = len(r.err)
}

func runScrape(db *sql.DB, mu *sync.Mutex, conf config.Config, scrapers ...scraper.Scraper) (Result, error) {
	mu.Lock()
	defer mu.Unlock()

	result := Result{}
	ctx := context.Background()
	if conf.HasScrapeTimeout() {
		tctx, cancel := context.WithTimeout(ctx, conf.ScrapeTimeout)
		ctx = tctx
		defer cancel()
	}

	queries := model.New(db)

	for _, scraper := range scrapers {
		slog.Info("running scraper", "scraper", scraper.Name())
		res := runScraper(ctx, scraper, queries)
		result.Add(res)
	}
	if len(result.err) > 0 {
		return result, errors.Join(result.err...)
	}
	return result, nil
}

func runScraper(ctx context.Context, scr scraper.Scraper, queries *model.Queries) Result {
	result := Result{}

	items, err := scr.FetchItems(ctx)
	if err != nil {
		result.err = append(result.err, err)
		return result
	}

	slog.Info("processig stories", "num", len(items))
	for _, itm := range items {

		slog.Debug("processig story", "story", itm)
		existing, err := queries.FindByScraperAndRef(ctx, model.FindByScraperAndRefParams{
			Scraper: itm.Scraper,
			RefID:   itm.RefID,
		})

		if err != nil {
			slog.Error("could not lookup story", "refid", itm.RefID, "scraper", itm.Scraper, "err", err)
			continue
		}

		if len(existing) > 0 {
			for _, story := range existing {
				updatedStory, err := queries.UpdateStory(ctx, model.UpdateStoryParams{
					Title:       itm.Title,
					Url:         itm.Url,
					Score:       itm.Score,
					NumComments: itm.NumComments,
					Type:        itm.Type,
					ID:          story.ID,
				})
				if err != nil {
					slog.Error("failed to updated story", "story", story, "err", err)
					result.err = append(result.err, err)
				} else {
					slog.Debug("updated existing story", "story", updatedStory)
					result.Updated++
				}
			}
			continue
		}

		story, err := queries.CreateStory(ctx, model.CreateStoryParams{
			RefID:       itm.RefID,
			Url:         itm.Url,
			By:          itm.By,
			PublishedAt: itm.PublishedAt,
			Title:       itm.Title,
			Type:        itm.Type,
			Score:       itm.Score,
			NumComments: itm.NumComments,
			Scraper:     itm.Scraper,
		})

		if err != nil {
			slog.Error("failed to create new story", "story", story, "err", err)
			result.err = append(result.err, err)
		} else {
			slog.Info("created new story", "story", story)
			result.New++
		}
	}

	// update recent stories that are not on the frontpage anymore
	now := time.Now()
	stories, err := queries.FindRecentForUpdate(ctx, model.FindRecentForUpdateParams{
		Scraper:   scr.Name(),
		UpdatedAt: now.Add(-15 * time.Minute),
		CreatedAt: now.Add(-24 * time.Hour),
	})

	if err != nil {
		slog.Error("failed to look up recent stories", "err", err)
		result.err = append(result.err, err)
		return result
	}

	slog.Info("updating recent stories", "num", len(stories))
	for _, story := range stories {
		itm, found, err := scr.FetchItem(ctx, story.RefID)
		if err != nil {
			slog.Error("failed to fetch recent story", "story", story, "err", err)
			result.err = append(result.err, err)
			continue
		}

		if !found {
			_, err := queries.MarkStoryDeleted(ctx, story.ID)
			if err != nil {
				slog.Error("failed to marked story deleted", "story", story, "err", err)
				result.err = append(result.err, err)
			} else {
				slog.Debug("marked story deleted", "story", story)
			}
			continue
		}

		slog.Debug("updating recent story", "story", story)
		updatedStory, err := queries.UpdateStory(ctx, model.UpdateStoryParams{
			Title:       itm.Title,
			Url:         itm.Url,
			Score:       itm.Score,
			NumComments: itm.NumComments,
			Type:        itm.Type,
			ID:          story.ID,
		})

		if err != nil {
			slog.Error("failed to update recent story", "story", story, "err", err)
			result.err = append(result.err, err)
		} else {
			result.Updated++
			slog.Debug("updated recent story", "story", updatedStory)
		}
	}
	return result
}
