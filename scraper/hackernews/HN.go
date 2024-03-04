package hackernews

import (
	"strings"
	"time"

	"github.com/floj/serializer-go/model"
)

type SearchResult struct {
	// Exhaustive          Exhaustive          `json:"exhaustive"`
	// ExhaustiveNbHits    bool                `json:"exhaustiveNbHits"`
	// ExhaustiveTypo      bool                `json:"exhaustiveTypo"`
	Hits []Hits `json:"hits"`
	// HitsPerPage         int                 `json:"hitsPerPage"`
	// NbHits              int                 `json:"nbHits"`
	// NbPages             int                 `json:"nbPages"`
	// Page                int                 `json:"page"`
	// Params              string              `json:"params"`
	ProcessingTimeMS int `json:"processingTimeMS"`
	// ProcessingTimingsMS ProcessingTimingsMS `json:"processingTimingsMS"`
	// Query               string              `json:"query"`
	ServerTimeMS int `json:"serverTimeMS"`
}

// type Exhaustive struct {
// 	NbHits bool `json:"nbHits"`
// 	Typo   bool `json:"typo"`
// }

// type Author struct {
// 	MatchLevel   string `json:"matchLevel"`
// 	MatchedWords []any  `json:"matchedWords"`
// 	Value        string `json:"value"`
// }

// type StoryText struct {
// 	MatchLevel   string `json:"matchLevel"`
// 	MatchedWords []any  `json:"matchedWords"`
// 	Value        string `json:"value"`
// }

// type Title struct {
// 	MatchLevel   string `json:"matchLevel"`
// 	MatchedWords []any  `json:"matchedWords"`
// 	Value        string `json:"value"`
// }

// type URL struct {
// 	MatchLevel   string `json:"matchLevel"`
// 	MatchedWords []any  `json:"matchedWords"`
// 	Value        string `json:"value"`
// }

// type HighlightResult struct {
// 	Author    Author    `json:"author"`
// 	StoryText StoryText `json:"story_text"`
// 	Title     Title     `json:"title"`
// 	URL       URL       `json:"url"`
// }

// type Request struct {
// 	RoundTrip int `json:"roundTrip"`
// }

// type Load struct {
// 	Dicts    int `json:"dicts"`
// 	Synonyms int `json:"synonyms"`
// 	Total    int `json:"total"`
// }

// type GetIdx struct {
// 	Load  Load `json:"load"`
// 	Total int  `json:"total"`
// }

//	type ProcessingTimingsMS struct {
//		Request Request `json:"_request"`
//		GetIdx  GetIdx  `json:"getIdx"`
//		Total   int     `json:"total"`
//	}

type Hits struct {
	// HighlightResult HighlightResult `json:"_highlightResult,omitempty"`
	Tags   []string `json:"_tags"`
	Author string   `json:"author"`
	// Children        []int           `json:"children,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedAtI  int       `json:"created_at_i"`
	NumComments int       `json:"num_comments,omitempty"`
	ObjectID    string    `json:"objectID"`
	Points      int       `json:"points,omitempty"`
	StoryID     int       `json:"story_id,omitempty"`
	StoryText   string    `json:"story_text,omitempty"`
	Title       string    `json:"title"`
	UpdatedAt   time.Time `json:"updated_at"`
	URL         string    `json:"url"`
}

func (h *Hits) GetType() string {
	check := []string{
		model.TypeHNAskHN,
		model.TypeHNShowHN,
		model.TypeHNJob,
		model.TypeHNStory,
	}
	for _, c := range check {
		if h.HasTag(c) {
			return c
		}
	}
	return model.TypeUnknown
}

func (h *Hits) HasTag(s string) bool {
	for _, t := range h.Tags {
		if strings.EqualFold(t, s) {
			return true
		}
	}
	return false
}
