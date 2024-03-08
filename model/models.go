// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package model

import (
	"time"
)

type Story struct {
	ID          int64
	RefID       string
	Url         string
	By          string
	PublishedAt time.Time
	UpdatedAt   time.Time
	CreatedAt   time.Time
	Title       string
	Type        string
	Score       int32
	NumComments int32
	Scraper     string
	Deleted     bool
}

type StoryHistory struct {
	ID        int64
	StoryID   int64
	Field     string
	OldVal    string
	NewVal    string
	CreatedAt time.Time
}
