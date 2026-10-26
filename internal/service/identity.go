package service

import "sync"

type Identity struct{ CreatorID, ModeratorID uint }

var once sync.Once
var identity *Identity

const creatorID uint = 1
const moderatorID uint = 2

// Lab 3 only: fixed identities. Replace with authenticated principals in Lab 4.
func CurrentUser() *Identity {
	once.Do(func() { identity = &Identity{CreatorID: creatorID, ModeratorID: moderatorID} })
	return identity
}
