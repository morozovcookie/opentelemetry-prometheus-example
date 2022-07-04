package v1

import (
	"net/url"
	"strconv"

	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
)

type SelfLink struct {
	Self string `json:"self"`
}

type Links struct {
	*SelfLink

	Base string `json:"base"`
	Next string `json:"next,omitempty"`
	Prev string `json:"prev,omitempty"`
}

func newLinks(baseURL *url.URL, pathPrefix string, opts otelexample.FindOptions, hasNext bool) (*Links, error) {
	links := new(Links)
	links.Base = baseURL.String()

	selfLink, err := baseURL.Parse(pathPrefix)
	if err != nil {
		return nil, err
	}

	links.SelfLink = &SelfLink{
		Self: selfLink.String(),
	}

	offset := opts.Offset()

	if !hasNext && offset == 0 {
		return links, nil
	}

	var (
		query = selfLink.Query()
		limit = opts.Limit()
	)

	if limit != otelexample.DefaultPageSize {
		query.Set("limit", strconv.FormatUint(limit, 10))
	}

	if offset > 0 && offset > limit {
		query.Set("start", strconv.FormatUint(offset-limit, 10))
	}

	if offset > 0 {
		selfLink.RawQuery = query.Encode()
		links.Prev = selfLink.RequestURI()
	}

	if hasNext {
		query.Set("start", strconv.FormatUint(offset+limit, 10))

		selfLink.RawQuery = query.Encode()
		links.Next = selfLink.RequestURI()
	}

	return links, nil
}
