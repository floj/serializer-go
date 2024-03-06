package config

import (
	"fmt"
	"time"
)

type Config struct {
	DBURI          string
	ScrapeInterval time.Duration
	CookieSecure   bool
}

func (c *Config) ScrapeEnabled() bool {
	return c.ScrapeInterval > time.Duration(0)
}

func Create(dbURI, scrape string, cookieSecure bool) (Config, error) {
	c := Config{
		DBURI:        dbURI,
		CookieSecure: cookieSecure,
	}
	d, err := time.ParseDuration(scrape)
	if err != nil {
		return c, fmt.Errorf("failed to parse scrape interval: %w", err)
	}
	c.ScrapeInterval = d.Abs()
	return c, nil
}
