package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var errPageExpired = errors.New("page expired")
var errPageOwner = errors.New("page belongs to another requester")

type pageState struct {
	owner   string
	pages   []string
	current int
	expires time.Time
}
type pageStore struct {
	mu      sync.Mutex
	entries map[string]pageState
	now     func() time.Time
	ttl     time.Duration
}

func newPageStore() *pageStore {
	return &pageStore{entries: map[string]pageState{}, now: time.Now, ttl: 15 * time.Minute}
}
func (p *pageStore) create(owner string, pages []string) (string, error) {
	b := make([]byte, 12)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	token := hex.EncodeToString(b)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.prune()
	p.entries[token] = pageState{owner: owner, pages: append([]string(nil), pages...), expires: p.now().Add(p.ttl)}
	return token, nil
}
func (p *pageStore) move(token, owner string, delta int) (string, int, int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.prune()
	s, ok := p.entries[token]
	if !ok {
		return "", 0, 0, errPageExpired
	}
	if owner != s.owner {
		return "", 0, 0, errPageOwner
	}
	s.current += delta
	if s.current < 0 {
		s.current = 0
	}
	if s.current >= len(s.pages) {
		s.current = len(s.pages) - 1
	}
	p.entries[token] = s
	return s.pages[s.current], s.current, len(s.pages), nil
}
func (p *pageStore) prune() {
	for k, v := range p.entries {
		if !p.now().Before(v.expires) {
			delete(p.entries, k)
		}
	}
}
