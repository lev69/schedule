package lib

import "errors"

var (
	ErrExist    = errors.New("already exists")
	ErrNotExist = errors.New("does not exist")
	ErrParse    = errors.New("parse error")
)
