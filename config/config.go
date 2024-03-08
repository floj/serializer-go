package config

import (
	"fmt"
	"time"
)

type Config struct {
	DBURI          string
	ScrapeInterval time.Duration
	ScrapeTimeout  time.Duration
	CookieSecure   bool
}

func (c *Config) ScrapeEnabled() bool {
	return c.ScrapeInterval > time.Duration(0)
}

func (c *Config) HasScrapeTimeout() bool {
	return c.ScrapeTimeout > time.Duration(0)
}

func Create(dbURI, scrapeInterval, scrapeTimeout string, cookieSecure bool) (Config, error) {
	c := Config{
		DBURI:        dbURI,
		CookieSecure: cookieSecure,
	}
	{
		d, err := time.ParseDuration(scrapeInterval)
		if err != nil {
			return c, fmt.Errorf("failed to parse scrape interval: %w", err)
		}
		c.ScrapeInterval = d.Abs()
	}
	{
		d, err := time.ParseDuration(scrapeTimeout)
		if err != nil {
			return c, fmt.Errorf("failed to parse scrape timeout: %w", err)
		}
		c.ScrapeTimeout = d.Abs()
	}
	return c, nil
}
