package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPagination(t *testing.T) {
	cases := []struct {
		payload []byte
		total   int
		next    *string
		err     error
	}{
		{
			payload: []byte(`{}`),
			total:   0,
			next:    nil,
			err:     nil,
		},
		{
			payload: []byte(`{"totalResultsCount": "12", "nextPageToken": null}`),
			total:   12,
			next:    nil,
			err:     nil,
		},
		{
			payload: []byte(`{"totalResultsCount": "12", "nextPageToken": "foo"}`),
			total:   12,
			next:    strPointer("foo"),
			err:     nil,
		},
		{
			payload: []byte(`{"totalResultsCount": "NaN"}`),
			total:   0,
			next:    nil,
			err:     fmt.Errorf("convert total to int: strconv.Atoi: parsing \"NaN\": invalid syntax"),
		},
	}

	for _, c := range cases {
		t.Run(string(c.payload), func(t *testing.T) {
			total, next, err := extractPaginationInfo(c.payload)
			assert.Equal(t, c.total, total)
			assert.Equal(t, c.next, next)
			if c.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, c.err.Error(), err.Error())
			}
		})
	}
}

func strPointer(s string) *string {
	return &s
}
