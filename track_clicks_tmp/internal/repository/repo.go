package repository

import (
	"context"
	"sync"
	"time"

	"example.com/dau/internal/date"
	"example.com/dau/internal/pkg"
)

type Repo struct {
	todaySets     map[int]map[int]struct{}
	yesterdaySets map[int]map[int]struct{}

	mutex *sync.RWMutex

	dateService date.Service
	lastUpdate  time.Time
}

func (r *Repo) Set(ctx context.Context, userID int, authorID int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	yesterday := r.dateService.Today().Add(-24 * time.Hour)

	if !r.lastUpdate.IsZero() && r.lastUpdate.After(pkg.DayStart(yesterday)) {
		r.yesterdaySets = r.todaySets
		r.todaySets = make(map[int]map[int]struct{})
	}

	if r.todaySets == nil {
		r.todaySets = make(map[int]map[int]struct{})
	}

	if r.todaySets[authorID] == nil {
		r.todaySets[authorID] = make(map[int]struct{})
	}

	r.todaySets[authorID][userID] = struct{}{}
	r.lastUpdate = now

	return nil
}

func (r *Repo) GetUniqueUsersCountForAuthors(ctx context.Context, authorIDs []int) ([]int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Если lastUpdate больше чем вчера, значит данные еще не "переехали" во вчера
	today := r.dateService.Today()
	yesterday := today.Add(-24 * time.Hour)

	if r.lastUpdate.After(pkg.DayStart(yesterday)) {
		// Данные еще не готовы для вчерашнего дня
		return []int{}, nil
	}

	var result []int

	for _, authorID := range authorIDs {
		if yesterdaySet, exists := r.yesterdaySets[authorID]; exists {
			result = append(result, len(yesterdaySet))
		}
	}

	return result, nil
}

func New(dateService date.Service) *Repo {
	return &Repo{
		todaySets:     make(map[int]map[int]struct{}),
		yesterdaySets: make(map[int]map[int]struct{}),
		mutex:         &sync.RWMutex{},
		dateService:   dateService,
	}
}
