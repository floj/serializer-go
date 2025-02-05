package config

import (
	"fmt"
	"time"
)

type DbConfig struct {
	Type         string
	RemoteURL    string
	AuthToken    string
	LocalPath    string
	SyncInterval time.Duration
}

type Config struct {
	DB             DbConfig
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

func CreateDB(dbUrl, authToken, localPath string) (DbConfig, error) {
	if dbUrl == "" {
		return DbConfig{}, fmt.Errorf("dbUrl must be set")
	}
	if authToken == "" {
		return DbConfig{}, fmt.Errorf("dbAuthToken must be set")
	}
	return DbConfig{
		Type:         "turso",
		RemoteURL:    dbUrl,
		AuthToken:    authToken,
		LocalPath:    localPath,
		SyncInterval: 15 * time.Minute,
	}, nil
}

func Create(db DbConfig, scrapeInterval, scrapeTimeout string, cookieSecure bool) (Config, error) {
	c := Config{
		DB:           db,
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
