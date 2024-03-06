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
					runScrape(db, mu, scrapers...)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return func(f func(Result, error) error) error {
			return f(runScrape(db, mu, scrapers...))
		}, func() {
			quit <- struct{}{}
		}
}

type Result struct {
	New     int
	Updated int
	Errors  int
}

func runScrape(db *sql.DB, mu *sync.Mutex, scrapers ...scraper.Scraper) (Result, error) {
	mu.Lock()
	defer mu.Unlock()

	result := Result{}
	ctx := context.Background()
	errs := []error{}
	for _, scraper := range scrapers {
		items, err := scraper.FetchItems(ctx)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		queries := model.New(db)

		for _, itm := range items {
			existing, err := queries.FindByScraperAndRef(ctx, model.FindByScraperAndRefParams{
				Scraper: itm.Scraper,
				RefID:   itm.RefID,
			})
			if err != nil {
				slog.Error("could not lookup story", "refid", itm.RefID, "scraper", itm.Scraper, "err", err)
				continue
			}
			if len(existing) > 0 {
				for _, s := range existing {
					s, err := queries.UpdateStory(ctx, model.UpdateStoryParams{
						Title:       itm.Title,
						Url:         itm.Url,
						Score:       itm.Score,
						NumComments: itm.NumComments,
						ID:          s.ID,
						Type:        itm.Type,
					})
					if err != nil {
						errs = append(errs, err)
						result.Errors++
						continue
					}
					result.Updated++
					slog.Info("updated story", "story", s)
				}
			} else {
				s, err := queries.CreateStory(ctx, model.CreateStoryParams{
					RefID:       itm.RefID,
					Url:         itm.Url,
					By:          itm.By,
					CreatedAt:   itm.CreatedAt,
					ScrapedAt:   itm.CreatedAt,
					Title:       itm.Title,
					Type:        itm.Type,
					Score:       itm.Score,
					NumComments: itm.NumComments,
					Scraper:     itm.Scraper,
				})
				if err != nil {
					errs = append(errs, err)
					result.Errors++
					continue
				}
				result.New++
				slog.Info("created new story", "story", s)
			}
		}

	}
	if len(errs) > 0 {
		return result, errors.Join(errs...)
	}
	return result, nil
}
