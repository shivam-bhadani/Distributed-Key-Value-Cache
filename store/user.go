package store

import "github.com/shivam-bhadani/distributed-cache/cache"

type Store struct {
	Username string
	Password string
	Cache    cache.ICache
}

func NewStore(username, password string) *Store {
	return &Store{
		Username: username,
		Password: password,
		Cache:    cache.NewCache(),
	}
}

func (s *Store) IsCorrectPassword(password string) bool {
	return s.Password == password
}
