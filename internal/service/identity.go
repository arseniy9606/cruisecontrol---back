package service

import (
	"errors"
	"sync"
)

var ErrInvalid = errors.New("invalid input or transition")

const CurrentUserID uint = 1

type Identity struct{ ID uint }

var once sync.Once
var current *Identity

func CurrentUser() *Identity {
	once.Do(func() { current = &Identity{ID: CurrentUserID} })
	return current
}
