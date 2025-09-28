package dau

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	events map[int]map[int]struct{} // authorID -> set of userIDs
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		events: make(map[int]map[int]struct{}),
	}
}

func (m *mockRepo) Set(ctx context.Context, userID int, authorID int) error {
	if m.events[authorID] == nil {
		m.events[authorID] = make(map[int]struct{})
	}
	m.events[authorID][userID] = struct{}{}
	return nil
}

func (m *mockRepo) GetAuthorList(ctx context.Context, userID int) ([]int, error) {
	var authors []int
	for authorID := range m.events {
		authors = append(authors, authorID)
	}
	return authors, nil
}

func (m *mockRepo) GetUniqueUsersCountForAuthors(ctx context.Context, authorIDs []int) ([]int, error) {
	var result []int
	for _, authorID := range authorIDs {
		if users, exists := m.events[authorID]; exists {
			result = append(result, len(users))
		}
		// Пропускаем авторов без данных
	}
	return result, nil
}

type mockDateService struct {
	today time.Time
}

func newMockDateService(today time.Time) *mockDateService {
	return &mockDateService{today: today}
}

func (m *mockDateService) Today() time.Time {
	return m.today
}

func TestService_Event(t *testing.T) {
	ctx := context.Background()
	dateService := newMockDateService(time.Now())
	repo := newMockRepo()
	service := NewService(dateService, repo)

	err := service.Event(ctx, &EventRequest{
		UserID:   123,
		AuthorID: 456,
	})
	require.NoError(t, err)

	authors, err := repo.GetAuthorList(ctx, 123)
	require.NoError(t, err)
	require.Len(t, authors, 1)
	require.Equal(t, 456, authors[0])
}

func TestService_Dau(t *testing.T) {
	ctx := context.Background()
	dateService := newMockDateService(time.Now())
	repo := newMockRepo()
	service := NewService(dateService, repo)

	events := []EventRequest{
		{UserID: 1, AuthorID: 100},
		{UserID: 2, AuthorID: 100},
		{UserID: 1, AuthorID: 200},
		{UserID: 3, AuthorID: 200},
		{UserID: 2, AuthorID: 200},
	}

	for _, event := range events {
		err := service.Event(ctx, &event)
		require.NoError(t, err)
	}

	authorIDs := []int{100, 200, 300}
	counts, err := service.Dau(ctx, authorIDs)
	require.NoError(t, err)
	require.Len(t, counts, 2) // Только авторы с данными

	// Проверяем, что есть данные для авторов 100 и 200
	require.Contains(t, counts, 2) // автор 100: 2 пользователя
	require.Contains(t, counts, 3) // автор 200: 3 пользователя
	// автор 300 пропущен, так как нет данных
}

func TestService_DuplicateUsers(t *testing.T) {
	ctx := context.Background()
	dateService := newMockDateService(time.Now())
	repo := newMockRepo()
	service := NewService(dateService, repo)

	event := EventRequest{UserID: 1, AuthorID: 100}

	err := service.Event(ctx, &event)
	require.NoError(t, err)

	err = service.Event(ctx, &event)
	require.NoError(t, err)

	counts, err := service.Dau(ctx, []int{100})
	require.NoError(t, err)
	require.Len(t, counts, 1)
	require.Equal(t, 1, counts[0])
}
