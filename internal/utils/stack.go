package utils

import (
	"container/list"
	"sync"
)

type Stack[T any] interface {
	Push(T)
	Pop() (T, bool)
	Peek() (T, bool)
	Len() int
}

type stackImpl[T any] struct {
	list  *list.List
	mutex sync.RWMutex
}

var _ Stack[int] = &stackImpl[int]{}

func NewStack[T any]() Stack[T] {
	return &stackImpl[T]{list: list.New()}
}

func (s *stackImpl[T]) Push(t T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.list.PushBack(t)
}

func (s *stackImpl[T]) Pop() (T, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	el := s.list.Back()
	if el == nil {
		var zero T
		return zero, false
	}
	s.list.Remove(el)
	return el.Value.(T), true
}

func (s *stackImpl[T]) Peek() (T, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	el := s.list.Back()
	if el == nil {
		var zero T
		return zero, false
	}
	return el.Value.(T), true
}

func (s *stackImpl[T]) Len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.list.Len()
}
