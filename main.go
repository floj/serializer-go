package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/floj/serializer-go/assets"
	"github.com/floj/serializer-go/config"
	"github.com/floj/serializer-go/job"
	"github.com/floj/serializer-go/model"
	"github.com/floj/serializer-go/scraper"
	"github.com/floj/serializer-go/scraper/hackernews"
	"github.com/floj/serializer-go/views"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lmittmann/tint"
)

func envOrDefault(name, def string) string {
	v := os.Getenv(name)
	if v != "" {
		return v
	}
	return def
}

func main() {
	//dbFile := flag.String("database-file", envOrDefault("DATABASE_FILE", "serializer.db"), "path to the database file")
	dbURL := flag.String("db-uri", envOrDefault("DB_URI", ""), "connection uri for the DB")
	scrapeInterval := flag.String("scrape-interval", envOrDefault("SCRAPE_INTERVAL", "1m"), "how often to scrape, set to 0 to disable scrape job")
	scrapeTimeout := flag.String("scrape-timeout", envOrDefault("SCRAPE_TIMEOUT", "1m"), "max time one scrape job is allowed to run set to 0 to for no limit")
	cookieInsecure := flag.Bool("cookie-insecure", false, "set secure flag on cookie")
	logLevel := flag.String("log-level", "info", "log level (debug, info, warn, error)")
	flag.Parse()

	tintOpts := &tint.Options{TimeFormat: time.RFC3339}
	switch *logLevel {
	case "debug":
		tintOpts.Level = slog.LevelDebug
	case "info":
		tintOpts.Level = slog.LevelInfo
	case "warn":
		tintOpts.Level = slog.LevelWarn
	case "error":
		tintOpts.Level = slog.LevelError
	default:
		panic("invalid log level: " + *logLevel)
	}

	conf, err := config.Create(*dbURL, *scrapeInterval, *scrapeTimeout, !*cookieInsecure)
	if err != nil {
		panic(err)
	}

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, tintOpts),
	))

	if err := run(conf); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func unreadCount(stories []model.Story, last int64) int {
	i := 0
	for _, s := range stories {
		if s.ID <= last {
			continue
		}
		i++
	}
	return i
}

func run(conf config.Config) error {
	if conf.DBURI == "" {
		return fmt.Errorf("db-uri is mandatory but was not set")
	}

	db, err := model.InitDB(conf)
	if err != nil {
		return fmt.Errorf("could not initialize DB connection: %w", err)
	}
	defer db.Close()

	scrapers, err := loadScrapers(conf)
	if err != nil {
		return err
	}

	trigger, stopJob := job.Start(db, conf, scrapers...)
	defer stopJob()

	if conf.ScrapeEnabled() {
		go trigger(func(r job.Result, err error) error {
			if err != nil {
				slog.Error("failed to scrape", "err", err)
				return nil
			}
			slog.Info("scrape success", "new", r.New, "updated", r.Updated)
			return nil
		})
	} else {
		slog.Info("scraping disabled")
	}

	app := echo.New()
	app.Use(
		middleware.Recover(),
		middleware.Logger(),
		middleware.Secure(),
	)

	app.StaticFS("/assets", assets.StaticAssets())

	app.GET("/", func(c echo.Context) error {
		last := getLastIdFromCookie(c.Cookie("serializer-go"))
		stories, err := getStories(c.Request().Context(), db, last)
		if err != nil {
			return err
		}
		unread := unreadCount(stories, last)
		return views.Index(stories, last, unread).Render(c.Request().Context(), c.Response())
	})

	app.GET("/clear", func(c echo.Context) error {
		writeCookie(c, 0, conf.CookieSecure)
		return c.Redirect(http.StatusSeeOther, "/")
	})

	app.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"status": "UP"})
	})

	app.POST("/", func(c echo.Context) error {
		last := getLastIdFromPOST(c)
		slog.Info("updating last", "last", last)
		writeCookie(c, last, conf.CookieSecure)
		return c.Redirect(http.StatusSeeOther, "/")
	})

	app.GET("/scrape", func(c echo.Context) error {
		return trigger(func(r job.Result, err error) error {
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"err": err,
				})
			}
			return c.JSON(http.StatusOK, r)
		})
	})

	return app.Start(":3000")
}

func getLastIdFromPOST(c echo.Context) int64 {
	p := c.FormValue("last")
	if p == "" {
		return 0
	}
	v, err := strconv.ParseInt(p, 10, 64)
	if err != nil {
		slog.Error("could not parse int", "err", err, "p", p)
		return 0
	}
	return v
}

type CookieVal struct {
	Last int64 `json:"last"`
}

func getLastIdFromCookie(c *http.Cookie, err error) int64 {
	if c == nil {
		return 0
	}
	if err != nil {
		slog.Error("could not parse cookie", "err", err, "cookie", c)
		return 0
	}
	v := CookieVal{}
	cv, err := url.QueryUnescape(c.Value)
	if err != nil {
		slog.Error("could not url decode cookie", "err", err, "cookie", c)
		return 0
	}
	err = json.Unmarshal([]byte(cv), &v)
	if err != nil {
		slog.Error("could not decode cookie value", "err", err, "cookie", c)
		return 0
	}
	return v.Last
}

func writeCookie(c echo.Context, last int64, secure bool) error {
	b := bytes.Buffer{}
	err := json.NewEncoder(&b).Encode(CookieVal{Last: last})
	if err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     "serializer-go",
		Value:    url.QueryEscape(b.String()),
		Expires:  time.Now().Add(time.Hour * 24 * 365),
		HttpOnly: true,
		Secure:   secure,
	})
	return nil
}

func getStories(ctx context.Context, db *sql.DB, id int64) ([]model.Story, error) {
	queries := model.New(db)
	stories, err := queries.ListStoriesBeginningAt(ctx, id-9)
	return stories, err
}

func loadScrapers(conf config.Config) ([]scraper.Scraper, error) {
	httpc := &http.Client{}
	hnScraper, err := hackernews.NewScraper(httpc)
	if err != nil {
		return nil, err
	}

	return []scraper.Scraper{
		hnScraper,
	}, nil
}
