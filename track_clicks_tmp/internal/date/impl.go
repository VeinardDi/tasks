package date

import "time"

type ServiceImpl struct {
}

func (s *ServiceImpl) Today() time.Time {
	return time.Now().Truncate(24 * time.Hour)
}

func NewService() *ServiceImpl {
	return new(ServiceImpl)
}
