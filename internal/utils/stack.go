package utils

import (
	"container/list"
	"sync"
)

type Stack[T any] interface {
	Push(*T)
	Pop() *T
	Peek() *T
	Len() int
}

type stackImpl[T interface{}] struct {
	list  *list.List
	mutex sync.RWMutex
}

var _ Stack[int] = &stackImpl[int]{}

func NewStack[T interface{}]() Stack[T] {
	return &stackImpl[T]{
		list: list.New(),
	}
}

func (s *stackImpl[T]) Push(t *T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.list.PushBack(t)
}

func (s *stackImpl[T]) Pop() *T {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	el := s.list.Back()
	if el == nil {
		return nil
	}
	s.list.Remove(el)
	return el.Value.(*T)
}

func (s *stackImpl[T]) Peek() *T {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	el := s.list.Back()
	if el == nil {
		return nil
	}
	return el.Value.(*T)
}

func (s *stackImpl[T]) Len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.list.Len()
}
