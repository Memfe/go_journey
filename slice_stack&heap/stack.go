package main

import (
	"errors"
)

type Stack struct {
	Items []int
}

func (s *Stack) Push(i int) {
	s.Items = append(s.Items, i)
}

func (s *Stack) Pop() (int, error) {
	if s.IsEmpty() {
		return 0, errors.New("Stack is empty")
	}

	l := len(s.Items) - 1
	toRemove := s.Items[l]

	s.Items[l] = 0
	s.Items = s.Items[:l]

	return toRemove, nil
}

func (s *Stack) Len() int {
	return len(s.Items)
}

func (s *Stack) IsEmpty() bool {
	return len(s.Items) == 0
}

func (s *Stack) Peek() (int, error) {
	if s.IsEmpty() {
		return 0, errors.New("Stack is empty")
	}
	l := len(s.Items) - 1
	top := s.Items[l]
	return top, nil
}

func (s *Stack) QueuePop() (int, error) {
	if s.IsEmpty() {
		return 0, errors.New("Queue is empty")
	}
	toRemove := s.Items[0]
	s.Items[0] = 0
	s.Items = s.Items[1:]
	return toRemove, nil
}
