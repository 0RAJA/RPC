package service

import "sync"

type UserStore interface {
	Save(user *User) error
	Find(username string) (*User, error)
}

type InMemoryUserStoreStore struct {
	mutex sync.RWMutex
	users map[string]*User
}

func (store *InMemoryUserStoreStore) Save(user *User) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if _, ok := store.users[user.Username]; ok {
		return ErrAlreadyExists
	}
	store.users[user.Username] = user
	return nil
}

func (store *InMemoryUserStoreStore) Find(username string) (*User, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	user := store.users[username]
	if user == nil {
		return nil, nil
	}
	return user.Clone(), nil
}

func NewInMemoryUserStoreStore() *InMemoryUserStoreStore {
	return &InMemoryUserStoreStore{
		mutex: sync.RWMutex{},
		users: make(map[string]*User),
	}
}
