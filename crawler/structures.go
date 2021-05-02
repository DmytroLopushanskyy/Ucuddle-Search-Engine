package main

import (
	"time"
)

var exists = struct{}{}

type Site struct {
	SiteId     uint64    `json:"site_id"`
	Title      string    `json:"title"`
	Link       string    `json:"link"`
	PageRank   uint64    `json:"page_rank"`
	Content    string    `json:"content"`
	Hyperlinks []string  `json:"hyperlinks"`
	AddedAt    time.Time `json:"added_at_time"`
}

type responseLinks struct {
	Links []string `json:"links"`
}

type set struct {
	dict map[string]struct{}
}

func NewSet() *set {
	s := &set{}
	s.dict = make(map[string]struct{})
	return s
}

func (s *set) Add(value string) {
	s.dict[value] = exists
}

func (s *set) Remove(value string) {
	delete(s.dict, value)
}

func (s *set) Contains(value string) bool {
	_, c := s.dict[value]
	return c
}
