package model

import (
	"net/url"
	"strings"
)

var urlTransformer = []URLTransformer{
	// NewFarsideTransformer("twitter.com", "nitter"),
	// NewFarsideTransformer("youtube.com", "piped"),
	// NewFarsideTransformer("reddit.com", "libreddit"),
	// NewFarsideTransformer("medium.com", "scribe"),
}

type URLTransformer interface {
	Matches(u *url.URL) bool
	Transform(u *url.URL) *url.URL
}

func NewFarsideTransformer(host, service string) URLTransformer {
	target, err := url.Parse("https://farside.link/" + service)
	if err != nil {
		panic(err)
	}
	return &farsideTransformer{
		host:           strings.ToLower(host),
		matchSubdomain: true,
		target:         *target,
	}
}

type farsideTransformer struct {
	host           string
	matchSubdomain bool
	target         url.URL
}

func (n *farsideTransformer) Matches(u *url.URL) bool {
	host := strings.ToLower(u.Hostname())
	if host == n.host {
		return true
	}
	if n.matchSubdomain && strings.HasSuffix(host, "."+n.host) {
		return true
	}
	return false
}

func (n *farsideTransformer) Transform(u *url.URL) *url.URL {
	u.Scheme = n.target.Scheme
	u.Host = n.target.Host
	u.Path = strings.TrimSuffix(n.target.Path, "/") + u.Path
	return u
}
