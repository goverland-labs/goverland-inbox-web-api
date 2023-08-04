package dao

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-web-api/internal/entities/dao"
)

const (
	daoCacheItemTTL    = 15 * time.Minute
	cleanCacheInterval = 5 * time.Minute
)

type cachedItem struct {
	expiresAt time.Time
	value     *dao.DAO
}

func (i cachedItem) expired() bool {
	return time.Now().After(i.expiresAt)
}

type Cache struct {
	mu    sync.RWMutex
	cache map[uuid.UUID]cachedItem
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
		cache: make(map[uuid.UUID]cachedItem),
	}

	go func() {
		for {
			<-time.After(cleanCacheInterval)

			repo.clean()
		}
	}()

	return repo
}

func (r *Cache) GetDaoByIDs(ids ...uuid.UUID) ([]*dao.DAO, []uuid.UUID) {
	hits := make([]*dao.DAO, 0, len(ids))
	missed := make([]uuid.UUID, 0, len(ids))
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

func (r *Cache) AddToCache(list ...*dao.DAO) {
	r.mu.Lock()
	defer r.mu.Unlock()

	expiresAt := time.Now().Add(daoCacheItemTTL)
	for i := range list {
		ci := cachedItem{
			expiresAt: expiresAt,
			value:     list[i],
		}
		r.cache[list[i].ID] = ci
	}
}

func (r *Cache) GetByID(id uuid.UUID) (*dao.DAO, bool) {
	r.mu.RLock()
	item, ok := r.cache[id]
	r.mu.RUnlock()

	if ok && !item.expired() {
		return item.value, true
	}

	return nil, false
}
