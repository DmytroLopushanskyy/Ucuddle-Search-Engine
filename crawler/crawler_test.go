package main

import (
	"strings"
	"testing"
)

// ============= Links for testing ukr detection and to change formats of the link to ukr language =============
//
// // ukr popular sites from here
//	// https://uk.wikipedia.org/wiki/%D0%A1%D0%BF%D0%B8%D1%81%D0%BE%D0%BA_%D0%BD%D0%B0%D0%B9%D0%BF%D0%BE%D0%BF%D1%83%D0%BB%D1%8F%D1%80%D0%BD%D1%96%D1%88%D0%B8%D1%85_%D1%96%D0%BD%D1%82%D0%B5%D1%80%D0%BD%D0%B5%D1%82-%D0%97%D0%9C%D0%86_%D0%B2_%D0%A3%D0%BA%D1%80%D0%B0%D1%97%D0%BD%D1%96
// 	links := append(res.Links[:5],
//		"https://www.pravda.com.ua/articles/2021/05/2/7292251/",
//		"https://www.segodnya.ua/",
//		"https://tsn.ua/",
//		"https://24tv.ua/",
//		"https://censor.net/",
//		"https://sport.ua/",
//		"https://korrespondent.net/",
//		"https://newsyou.info/",
//		"https://from-ua.com/",
//		"https://www.rada.gov.ua/",
//		"https://www.ukr.net/",
//		)
//
//	// test social networks
//	links = append(links,
//	"https://uk-ua.facebook.com/login/web/",
//		"https://twitter.com/login/?lang=uk",
//		"https://twitter.com/login/?lang=ru",
//		"https://www.facebook.com/login/web/",
//		"https://www.instagram.com/",
//		"https://twitter.com/",
//	)
//
//	// test sites of weather forecast
//	links = append(links,
//	"https://ua.sinoptik.ua/%D0%BF%D0%BE%D0%B3%D0%BE%D0%B4%D0%B0-%D0%BB%D1%8C%D0%B2%D1%96%D0%B2",
//		"https://sinoptik.ua/",
//		"https://www.gismeteo.ua/ua/weather-lviv-4949/",
//		"https://rp5.ua/%D0%9F%D0%BE%D0%B3%D0%BE%D0%B4%D0%B0_%D1%83_%D0%9B%D1%8C%D0%B2%D0%BE%D0%B2%D1%96_(%D0%B0%D0%B5%D1%80%D0%BE%D0%BF%D0%BE%D1%80%D1%82)",
//	)
//
//	// test other sites
//	links = append(links,
//		"https://www.google.com/",
//		"https://www.youtube.com/",
//		"https://www.olx.ua/",
//		"https://rozetka.com.ua/",
//		"https://www.work.ua/",
//		"https://tabletki.ua/",
//		"https://www.ria.com/",
//		"https://rst.ua/ukr/",
//		"https://en.wikipedia.org/wiki/%D0%93%D0%BE%D0%BB%D0%BE%D0%B2%D0%BD%D0%B0_%D1%81%D1%82%D0%BE%D1%80%D1%96%D0%BD%D0%BA%D0%B0",
//	)

// problem links for testing
// https://spacenews.com.ua/op-ed-in-defense-of-regulation//

// EqualArrays tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func EqualArrays(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestGetAllSiteUkrLinks1(t *testing.T) {
	got := getAllSiteUkrLinks("https://www.example.com/")
	expected := []string{"https://www.example.com/", "https://www.example.com/uk/",
		"https://www.example.com/?lang=uk/", "https://www.example.ua/", "https://uk.example.com/",
		"https://ua.example.com/", "https://www.example.com.ua/", "https://uk-ua.example.com/"}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}

func TestGetAllSiteUkrLinks2(t *testing.T) {
	got := getAllSiteUkrLinks("https://example.com/")
	expected := []string{"https://example.com/", "https://example.com/uk/", "https://example.com/?lang=uk/",
		"https://example.ua/", "https://ua.example.com/", "https://uk.example.com/", "https://example.com.ua/"}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}

func TestGetAllSiteUkrLinks3(t *testing.T) {
	got := getAllSiteUkrLinks("https://example/")
	expected := []string{"https://example/", "https://example/uk/", "https://example/?lang=uk/",
		"https://ua.example/", "https://uk.example/"}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}

func TestGetAllSiteUkrLinks4(t *testing.T) {
	got := getAllSiteUkrLinks("https://example")
	expected := []string{"https://example/", "https://example/uk/", "https://example/?lang=uk/",
		"https://ua.example/", "https://uk.example/"}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}

func TestGetAllSiteUkrLinksRealLink(t *testing.T) {
	got := getAllSiteUkrLinks("https://onlinecorrector.com.ua/")
	expected := []string{"https://onlinecorrector.com.ua/", "https://onlinecorrector.com.ua/uk/",
		"https://onlinecorrector.com.ua/?lang=uk/"}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}

func TestGetAllSiteUkrLinksRealLink2(t *testing.T) {
	got := getAllSiteUkrLinks("https://en.wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0")
	expected := []string{"https://en.wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/",
		"https://wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/uk/",
		"https://wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/?lang=uk/",
		"https://wikipedia.ua/",
		"https://ua.wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/",
		"https://uk.wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/"}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}
