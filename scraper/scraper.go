package scraper

import (
	"context"

	"github.com/floj/serializer-go/model"
)

type Scraper interface {
	Name() string
	FetchItem(ctx context.Context, refId string) (model.Story, bool, error)
	FetchItems(ctx context.Context) ([]model.Story, error)
}
