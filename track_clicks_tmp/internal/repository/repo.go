package repository

import (
	"context"
	"example.com/dau/internal/date"
	"example.com/dau/internal/pkg"
	"sync"
	"time"
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
	today := r.dateService.Today()

	if !r.lastUpdate.IsZero() && pkg.DayStart(today).After(pkg.DayStart(r.lastUpdate)) {
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
