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
	err     []error
}

func (r *Result) Add(other Result) {
	r.New += other.New
	r.Updated += other.Updated
	r.err = append(r.err, other.err...)
	r.Errors = len(r.err)
}

func runScrape(db *sql.DB, mu *sync.Mutex, scrapers ...scraper.Scraper) (Result, error) {
	mu.Lock()
	defer mu.Unlock()

	result := Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	queries := model.New(db)

	for _, scraper := range scrapers {
		res := runScraper(ctx, scraper, queries)
		result.Add(res)
	}
	if len(result.err) > 0 {
		return result, errors.Join(result.err...)
	}
	return result, nil
}

func runScraper(ctx context.Context, scr scraper.Scraper, queries *model.Queries) Result {
	scrapeTime := time.Now()
	result := Result{}

	items, err := scr.FetchItems(ctx)
	if err != nil {
		result.err = append(result.err, err)
		return result
	}

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
			for _, story := range existing {

				diff := story.Diff(itm)

				story, err = queries.UpdateStory(ctx, model.UpdateStoryParams{
					Title:       itm.Title,
					Url:         itm.Url,
					Score:       itm.Score,
					NumComments: itm.NumComments,
					ID:          story.ID,
					UpdatedAt:   scrapeTime,
					Type:        itm.Type,
				})
				if err != nil {
					result.err = append(result.err, err)
					continue
				}

				for k, v := range diff {
					queries.CreateStoryHistory(ctx, model.CreateStoryHistoryParams{
						StoryID:   story.ID,
						Field:     k,
						OldVal:    v.Old,
						NewVal:    v.New,
						CreatedAt: scrapeTime,
					})
				}
				result.Updated++
				slog.Info("updated story", "story", story)
			}
		} else {
			story, err := queries.CreateStory(ctx, model.CreateStoryParams{
				RefID:       itm.RefID,
				Url:         itm.Url,
				By:          itm.By,
				CreatedAt:   itm.CreatedAt,
				ScrapedAt:   scrapeTime,
				UpdatedAt:   scrapeTime,
				Title:       itm.Title,
				Type:        itm.Type,
				Score:       itm.Score,
				NumComments: itm.NumComments,
				Scraper:     itm.Scraper,
			})
			if err != nil {
				result.err = append(result.err, err)
				continue
			}
			result.New++
			slog.Info("created new story", "story", story)
		}
	}

	// update recent stories that are not on the frontpage anymore
	stories, err := queries.FindRecentForUpdate(ctx, model.FindRecentForUpdateParams{
		Scraper:   scr.Name(),
		UpdatedAt: scrapeTime.Add(-15 * time.Minute),
		CreatedAt: scrapeTime.Add(-24 * time.Hour),
	})
	if err != nil {
		result.err = append(result.err, err)
		return result
	}
	slog.Info("updading recent stories", "num", len(stories))

	for _, story := range stories {
		itm, err := scr.FetchItem(ctx, story.RefID)
		if err != nil {
			result.err = append(result.err, err)
			continue
		}
		diff := story.Diff(itm)

		story, err := queries.UpdateStory(ctx, model.UpdateStoryParams{
			Title:       itm.Title,
			Url:         itm.Url,
			Score:       itm.Score,
			NumComments: itm.NumComments,
			ID:          story.ID,
			UpdatedAt:   scrapeTime,
			Type:        itm.Type,
		})
		if err != nil {
			result.err = append(result.err, err)
			continue
		}

		for k, v := range diff {
			queries.CreateStoryHistory(ctx, model.CreateStoryHistoryParams{
				StoryID:   story.ID,
				Field:     k,
				OldVal:    v.Old,
				NewVal:    v.New,
				CreatedAt: scrapeTime,
			})
		}
		result.Updated++
		slog.Info("updated story", "story", story)
	}
	return result
}
