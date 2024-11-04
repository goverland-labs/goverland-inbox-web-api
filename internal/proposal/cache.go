package proposal

import (
	"sync"
	"time"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/proposal"
)

const (
	proposalCacheItemTTL = 5 * time.Minute
	cleanCacheInterval   = 1 * time.Minute
)

type cachedItem struct {
	expiresAt time.Time
	value     *proposal.Proposal
}

func (i cachedItem) expired() bool {
	return time.Now().After(i.expiresAt)
}

type Cache struct {
	mu    sync.RWMutex
	cache map[string]cachedItem
}

func (r *Cache) clean() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id := range r.cache {
		if !r.cache[id].expired() {
			continue
		}

		delete(r.cache, id)
	}
}

func NewCache() *Cache {
	repo := &Cache{
		cache: make(map[string]cachedItem),
	}

	go func() {
		for {
			<-time.After(cleanCacheInterval)

			repo.clean()
		}
	}()

	return repo
}

func (r *Cache) GetProposalsByIDs(ids ...string) ([]*proposal.Proposal, []string) {
	hits := make([]*proposal.Proposal, 0, len(ids))
	missed := make([]string, 0, len(ids))
	r.mu.RLock()
	for _, id := range ids {
		item, ok := r.cache[id]
		if ok && !item.expired() {
			hits = append(hits, item.value)
		} else {
			missed = append(missed, id)
		}
	}
	r.mu.RUnlock()

	return hits, missed
}

func (r *Cache) AddToCache(list ...*proposal.Proposal) {
	r.mu.Lock()
	defer r.mu.Unlock()

	expiresAt := time.Now().Add(proposalCacheItemTTL)
	for i := range list {
		ci := cachedItem{
			expiresAt: expiresAt,
			value:     list[i],
		}
		r.cache[list[i].ID] = ci
	}
}

func (r *Cache) GetByID(id string) (*proposal.Proposal, bool) {
	r.mu.RLock()
	item, ok := r.cache[id]
	r.mu.RUnlock()

	if ok && !item.expired() {
		return item.value, true
	}

	return nil, false
}
