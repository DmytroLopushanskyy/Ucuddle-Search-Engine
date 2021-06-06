package main

import (
	"sync"
)

type SafeSetOfLinks struct {
	mu      sync.Mutex
	linkSet map[string]struct{}
}

type SafeListOfSites struct {
	mu          sync.Mutex
	actualSites []Site
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func newSafeSetOfLinks() *SafeSetOfLinks {
	s := &SafeSetOfLinks{}
	s.linkSet = make(map[string]struct{})
	return s
}

func (c *SafeSetOfLinks) addLink(link *string) {
	c.mu.Lock()

	// Lock so only one goroutine at a time can access the map c.v.
	c.linkSet[*link] = exists
	c.mu.Unlock()
}

func (c *SafeSetOfLinks) checkIfContains(link *string) bool {
	c.mu.Lock()
	_, isFound := c.linkSet[*link]
	c.mu.Unlock()
	return isFound
}

func (c *SafeListOfSites) addSite(site *Site) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.actualSites = append(c.actualSites, *site)
	c.mu.Unlock()
}

func (c *SafeListOfSites) checkIfContains(site *Site) (int, bool) {
	c.mu.Lock()
	for i, item := range c.actualSites {
		if item.Link == (*site).Link {
			c.mu.Unlock()
			return i, true
		}
	}
	c.mu.Unlock()
	return -1, false
}
