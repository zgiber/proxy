package auth

import (
	"errors"
	"sync"
)

type TokenStore interface {
	Put(pub, priv string)
	Get(pub string) (string, bool)
	Delete(pub string)
}

type InMemTokenStore struct {
	sync.RWMutex
	tokens map[string]string
}

func (ms *InMemTokenStore) Put(pub, priv string) {
	ms.Lock()
	ms.tokens[pub] = priv
	ms.Unlock()
}

func (ms *InMemTokenStore) Get(pub string) (string, bool) {
	ms.RLock()
	priv, ok := ms.tokens[pub]
	ms.RUnlock()
	return priv, ok
}

func (ms *InMemTokenStore) Delete(pub, priv string) {
	ms.Lock()
	delete(ms.tokens, pub)
	ms.Unlock()
}

func (ms *InMemTokenStore) Exchange(pub string) (string, error) {
	priv, ok := ms.Get(pub)
	if !ok {
		return "", errors.New("invalid token")
	}
	return priv, nil
}
