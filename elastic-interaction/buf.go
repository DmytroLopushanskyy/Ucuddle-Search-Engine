package main

import (
	"strings"
	"sync"
)


var exists = struct{}{}

type SafeSetOfLinks struct {
	mu      sync.Mutex
	linkSet map[string]struct{}
}

func newSafeSetOfLinks() *SafeSetOfLinks {
	s := &SafeSetOfLinks{}
	s.linkSet = make(map[string]struct{})
	return s
}

func (c *SafeSetOfLinks) addLink(link string) {
	c.mu.Lock()

	// Lock so only one goroutine at a time can access the map c.v.
	c.linkSet[link] = exists
	c.mu.Unlock()
}

func (c *SafeSetOfLinks) checkIfContains(link string) bool {
	c.mu.Lock()
	_, isFound := c.linkSet[link]
	c.mu.Unlock()
	return isFound
}

func findNthSymbol(content string, symbol string, nOccurrences int) int {
	startPos := 0
	endPos := len(content)
	pos := -1
	for i := 0; i < nOccurrences; i++ {
		pos = strings.Index(content[startPos: endPos], symbol)
		if pos == -1 {
			return -1
		}

		startPos += pos + 1
	}

	return startPos - 1
}
