package scraper

import (
	"context"

	"github.com/floj/serializer-go/model"
)

type Scraper interface {
	FetchItems(ctx context.Context) ([]model.Story, error)
}
