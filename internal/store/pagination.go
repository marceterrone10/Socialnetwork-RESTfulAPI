package store

import (
	"net/http"
	"strconv"
)

type PaginatedQuery struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=10"`
	Offset int    `json:"offset" validate:"gte=0"`
	Sort   string `json:"sort" validate:"oneof=asc desc"`
}

func (q *PaginatedQuery) ParseURLParams(r *http.Request) (PaginatedQuery, error) {
	qs := r.URL.Query()

	limit := qs.Get("limit")
	if limit != "" {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			return *q, err
		}

		q.Limit = limitInt // return saved value for limit
	}

	offset := qs.Get("offset")
	if offset != "" {
		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			return *q, err
		}

		q.Offset = offsetInt // return saved value for offset
	}

	sort := qs.Get("sort")
	if sort != "" {
		q.Sort = sort
	}

	return *q, nil
}
