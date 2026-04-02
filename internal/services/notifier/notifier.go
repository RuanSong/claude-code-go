package notifier

import (
	"sync"
	"time"
)

type NotificationType string

const (
	TypeInfo    NotificationType = "info"
	TypeSuccess NotificationType = "success"
	TypeWarning NotificationType = "warning"
	TypeError   NotificationType = "error"
)

type Notification struct {
	ID        string
	Type      NotificationType
	Title     string
	Message   string
	Timestamp time.Time
	Dismissed bool
}

type NotifierService struct {
	mu               sync.RWMutex
	notifications    []*Notification
	listeners        []chan *Notification
	maxNotifications int
}

var (
	instance     *NotifierService
	instanceOnce sync.Once
)

func GetNotifierService() *NotifierService {
	instanceOnce.Do(func() {
		instance = &NotifierService{
			notifications:    make([]*Notification, 0),
			listeners:        make([]chan *Notification, 0),
			maxNotifications: 100,
		}
	})
	return instance
}

func (s *NotifierService) Notify(notifType NotificationType, title, message string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()
	notif := &Notification{
		ID:        id,
		Type:      notifType,
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
	}

	s.notifications = append(s.notifications, notif)

	if len(s.notifications) > s.maxNotifications {
		s.notifications = s.notifications[1:]
	}

	for _, ch := range s.listeners {
		select {
		case ch <- notif:
		default:
		}
	}

	return id
}

func (s *NotifierService) Info(title, message string) string {
	return s.Notify(TypeInfo, title, message)
}

func (s *NotifierService) Success(title, message string) string {
	return s.Notify(TypeSuccess, title, message)
}

func (s *NotifierService) Warning(title, message string) string {
	return s.Notify(TypeWarning, title, message)
}

func (s *NotifierService) Error(title, message string) string {
	return s.Notify(TypeError, title, message)
}

func (s *NotifierService) GetNotifications() []*Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Notification, len(s.notifications))
	copy(result, s.notifications)
	return result
}

func (s *NotifierService) GetActiveNotifications() []*Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Notification
	for _, n := range s.notifications {
		if !n.Dismissed {
			result = append(result, n)
		}
	}
	return result
}

func (s *NotifierService) Dismiss(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, n := range s.notifications {
		if n.ID == id {
			n.Dismissed = true
			return true
		}
	}
	return false
}

func (s *NotifierService) DismissAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, n := range s.notifications {
		n.Dismissed = true
	}
}

func (s *NotifierService) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.notifications = make([]*Notification, 0)
}

func (s *NotifierService) Subscribe() chan *Notification {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *Notification, 10)
	s.listeners = append(s.listeners, ch)
	return ch
}

func (s *NotifierService) Unsubscribe(ch chan *Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, listener := range s.listeners {
		if listener == ch {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			close(ch)
			break
		}
	}
}

func (s *NotifierService) GetUnreadCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, n := range s.notifications {
		if !n.Dismissed {
			count++
		}
	}
	return count
}

func (s *NotifierService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total":     len(s.notifications),
		"active":    s.GetUnreadCount(),
		"listeners": len(s.listeners),
	}

	byType := make(map[NotificationType]int)
	for _, n := range s.notifications {
		byType[n.Type]++
	}
	stats["by_type"] = byType

	return stats
}

func generateID() string {
	return time.Now().Format("20060102150405.000000")
}
