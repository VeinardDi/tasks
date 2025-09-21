package pkg

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Store хранит множества userID по authorID для сегодня/вчера
type Store struct {
	clk Clock

	mu sync.RWMutex

	todayDay     time.Time
	yesterdayDay time.Time

	// authorID -> set(userID)
	today     map[string]map[string]struct{}
	yesterday map[string]map[string]struct{}
}

func NewStore(clk Clock) *Store {
	if clk == nil {
		clk = SystemClock{}
	}
	now := clk.Now()
	today := dayStart(now)
	yesterdayDay := today.AddDate(0, 0, -1)

	return &Store{
		clk:          clk,
		todayDay:     today,
		yesterdayDay: yesterdayDay,
		today:        make(map[string]map[string]struct{}),
		yesterday:    make(map[string]map[string]struct{}),
	}
}

func (s *Store) AddClick(ctx context.Context, authorID, userID string) error {
	if authorID == "" || userID == "" {
		return errors.New("authorID and userID must be non-empty")
	}

	now := s.clk.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.rotateIfNeeded(now)

	set := s.today[authorID]
	if set == nil {
		set = make(map[string]struct{}, 1)
		s.today[authorID] = set
	}
	set[userID] = struct{}{}
	return nil
}

func (s *Store) CountYesterday(ctx context.Context, authorIDs []string) map[string]int64 {
	now := s.clk.Now()

	s.mu.Lock()
	s.rotateIfNeeded(now)
	s.mu.Unlock()

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]int64, len(authorIDs))
	for _, a := range authorIDs {
		if a == "" {
			continue
		}
		if set := s.yesterday[a]; set != nil {
			out[a] = int64(len(set))
		} else {
			out[a] = 0
		}
	}
	return out
}

func (s *Store) rotateIfNeeded(now time.Time) {
	todayNow := dayStart(now)
	if !todayNow.Equal(s.todayDay) {

		s.yesterday = s.today
		s.yesterdayDay = s.todayDay

		s.today = make(map[string]map[string]struct{})
		s.todayDay = todayNow
	}
}

func dayStart(t time.Time) time.Time {
	lt := t.In(time.Local)
	y, m, d := lt.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}
