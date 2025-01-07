package libsql

import (
	"github.com/go-on-bike/bike/assert"
)

type options struct {
	url *string
}

type Option func(options *options)

func WithURL(url string) Option {
	return func(options *options) {
		assert.NotEmptyString(url, "URL options is empty in libsql operator")
		options.url = &url
	}
}
