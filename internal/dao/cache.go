package dao

import (
	"strings"
	"sync"
	"time"

	"github.com/goverland-labs/goverland-inbox-web-api/internal/entities/dao"
)

const (
	daoCacheItemTTL    = 5 * time.Minute
	cleanCacheInterval = 1 * time.Minute
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

func (r *Cache) GetDaoByIDs(ids ...string) (map[string]*dao.DAO, []string) {
	hits := make(map[string]*dao.DAO, len(ids))
	missed := make([]string, 0, len(ids))
	r.mu.RLock()
	for _, id := range ids {
		lowered := strings.ToLower(id)
		item, ok := r.cache[lowered]
		if ok && !item.expired() {
			hits[strings.ToLower(item.value.ID.String())] = item.value
		} else {
			missed = append(missed, lowered)
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

		// add to cache by alias and internal IDs
		r.cache[strings.ToLower(list[i].ID.String())] = ci
		r.cache[strings.ToLower(list[i].Alias)] = ci
	}
}

func (r *Cache) GetByID(id string) (*dao.DAO, bool) {
	r.mu.RLock()
	item, ok := r.cache[strings.ToLower(id)]
	r.mu.RUnlock()

	if ok && !item.expired() {
		return item.value, true
	}

	return nil, false
}
