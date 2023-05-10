package redis

import (
	"errors"
)

var ErrUnexpectedCachedResource = errors.New("unexpected cached resource") // received cache resource was not expected
