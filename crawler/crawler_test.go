package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"strings"
	"testing"
)


// ============= Links for testing ukr detection and to change formats of the link to ukr language =============
//
//	----------------- TEST 1. Sites on Ukrainian -----------------
//
// // ukr popular sites from here
// // test news sites, which are real and ukrainian
//	links := append(res.Links[:10],
//		"https://www.pravda.com.ua/articles/2021/05/2/7292251/",
//		"https://tsn.ua/",
//		"https://24tv.ua/",
//		"https://from-ua.com/",
//		"https://www.rada.gov.ua/",
//		"https://www.ukr.net/",
//	)
//
//	// test true social networks
//	links = append(links,
//		"https://uk-ua.facebook.com/login/web/",
//		"https://twitter.com/login/?lang=uk",
//	)
//
//	// test sites of weather forecast
//	links = append(links,
//		"https://ua.sinoptik.ua/%D0%BF%D0%BE%D0%B3%D0%BE%D0%B4%D0%B0-%D0%BB%D1%8C%D0%B2%D1%96%D0%B2",
//		"https://www.gismeteo.ua/ua/weather-lviv-4949/",
//		"https://rp5.ua/%D0%9F%D0%BE%D0%B3%D0%BE%D0%B4%D0%B0_%D1%83_%D0%9B%D1%8C%D0%B2%D0%BE%D0%B2%D1%96_(%D0%B0%D0%B5%D1%80%D0%BE%D0%BF%D0%BE%D1%80%D1%82)",
//	)

// ================= Russian sites =================

//links = append(links,
//	"https://from-ua.com/",
//	"https://www.segodnya.ua/",
//	"https://censor.net/",
//	"https://korrespondent.net/",
//	"https://newsyou.info/",
//	"https://www.facebook.com/login/web/",
//	"https://www.instagram.com/",
//	"https://www.olx.ua/",
//	"https://rozetka.com.ua/",
//)
//

// ================= Should find sites =================

//links = append(links,
//	"https://sport.ua/", // should find https://sport.ua/uk
//	"https://twitter.com/login/", // should find https://twitter.com/login?lang=uk
//	"https://sinoptik.ua/", // https://ua.sinoptik.ua/
//	"https://www.google.com/", // should find https://www.google.com.ua/
//)
//

// ================= Hard detected ukr sites =================

//links := append(res.Links[:10],
//"https://twitter.com/login/?lang=uk",
//"https://rp5.ua/%D0%9F%D0%BE%D0%B3%D0%BE%D0%B4%D0%B0_%D1%83_%D0%9B%D1%8C%D0%B2%D0%BE%D0%B2%D1%96_(%D0%B0%D0%B5%D1%80%D0%BE%D0%BF%D0%BE%D1%80%D1%82)",
//"https://rst.ua/ukr/",
//)

//// hard sites
//links = append(links,
//	"https://www.youtube.com/",
//)
//
//// sites with other languages
//links = append(links,
//	"https://en.wikipedia.org/wiki/%D0%93%D0%BE%D0%BB%D0%BE%D0%B2%D0%BD%D0%B0_%D1%81%D1%82%D0%BE%D1%80%D1%96%D0%BD%D0%BA%D0%B0",
//)

// ================= Should NOT find sites ================= .....

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

func setUpCollector(collector *colly.Collector, site *Site, mum *map[string][]string,
	hyperlinksSet *set) {
	collector.OnRequest(func(request *colly.Request) {
		standardLogger.Println("Visiting", request.URL.String())
	})

	collector.OnResponse(func(response *colly.Response) {
		if response.StatusCode != 200 {
			standardLogger.Println(response.StatusCode)
		}
	})

	collector.OnError(func(response *colly.Response, err error) {
		if response.StatusCode != 200 && response.StatusCode != 0 {
			standardLogger.Println(response.StatusCode)
		}
	})

	collector.OnHTML("p", func(element *colly.HTMLElement) {
		(*mum)[element.Name] = append((*mum)[element.Name], element.Text)
		site.Link = strings.TrimSpace((element.Request).URL.String())
	})

	collector.OnHTML("div", func(element *colly.HTMLElement) {
		(*mum)[element.Name] = append((*mum)[element.Name], element.Text)
		site.Link = strings.TrimSpace((element.Request).URL.String())
	})

	collector.OnHTML("li", func(element *colly.HTMLElement) {
		(*mum)[element.Name] = append((*mum)[element.Name], element.Text)
		site.Link = strings.TrimSpace((element.Request).URL.String())
	})

	collector.OnHTML("article", func(element *colly.HTMLElement) {
		(*mum)[element.Name] = append((*mum)[element.Name], element.Text)
		site.Link = strings.TrimSpace((element.Request).URL.String())
	})

	collector.OnHTML("head", func(element *colly.HTMLElement) {
		site.Link = strings.TrimSpace((element.Request).URL.String())

		if len(site.Title) < 3 {
			site.Title = element.ChildText("title")
		}

		if len(site.Title) < 3 {
			site.Title = element.DOM.Find("title").Text()
		}
	})

	collector.OnHTML("title", func(element *colly.HTMLElement) {
		if len(site.Title) < 3 {
			site.Title = element.Text
		}
	})

	collector.OnHTML("h1", func(element *colly.HTMLElement) {
		if len(site.Title) < 3 {
			site.Title = element.Text
		}
	})

	collector.OnHTML("html", func(e *colly.HTMLElement) {
		if len(site.Title) < 3 {
			e.ChildAttr(`meta[property="og:title"]`, "content")
		}

		if len(site.Title) < 3 {
			e.ChildText("title")
		}

		if len(site.Title) < 3 {
			e.DOM.Find("title").Text()
		}
	})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := strings.TrimSpace(e.Request.AbsoluteURL(e.Attr("href")))
		if link != "" && len(link) > 5 && link[:5] == "https" {
			// clear from link parameters
			startLinkParameters := strings.Index(link, "?")

			if startLinkParameters > 0 {
				link = link[:startLinkParameters]
			}

			if link[len(link)-1:len(link)] == "/" {
				link = link[:len(link)-1]
			}

			hyperlinksSet.Add(&link)
		}
	})
}

func TestGetAllSiteUkrLinks1(t *testing.T) {
	got := getAllSiteUkrLinks("https://www.example.com/")
	expected := []string{"https://www.example.com/", "https://www.example.com/uk/", "https://www.example.com/ua/", "https://www.example.com/ukr/",
		"https://www.example.com/?lang=uk/", "https://www.example.ua/", "https://uk.example.com/",
		"https://ua.example.com/", "https://www.example.com.ua/", "https://uk-ua.example.com/"}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}

func TestGetAllSiteUkrLinks2(t *testing.T) {
	got := getAllSiteUkrLinks("https://example.com/")
	expected := []string{"https://example.com/", "https://example.com/uk/", "https://example.com/ua/", "https://example.com/ukr/",
		"https://example.com/?lang=uk/", "https://example.ua/", "https://ua.example.com/",
		"https://uk.example.com/", "https://example.com.ua/"}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}

func TestGetAllSiteUkrLinks3(t *testing.T) {
	got := getAllSiteUkrLinks("https://example/")
	expected := []string{"https://example/", "https://example/uk/", "https://example/ua/", "https://example/ukr/",
		"https://example/?lang=uk/", "https://ua.example/", "https://uk.example/"}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}

func TestGetAllSiteUkrLinks4(t *testing.T) {
	got := getAllSiteUkrLinks("https://example")
	expected := []string{"https://example/", "https://example/uk/", "https://example/ua/", "https://example/ukr/",
		"https://example/?lang=uk/", "https://ua.example/", "https://uk.example/"}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}

func TestGetAllSiteUkrLinksRealLink(t *testing.T) {
	got := getAllSiteUkrLinks("https://onlinecorrector.com.ua/")
	expected := []string{"https://onlinecorrector.com.ua/", "https://onlinecorrector.com.ua/uk/",
		"https://onlinecorrector.com.ua/ua/", "https://onlinecorrector.com.ua/ukr/", "https://onlinecorrector.com.ua/?lang=uk/"}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}

func TestGetAllSiteUkrLinksRealLink2(t *testing.T) {
	got := getAllSiteUkrLinks("https://en.wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0")
	expected := []string{"https://en.wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/",
		"https://wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/uk/",
		"https://wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/ua/",
		"https://wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/ukr/",
		"https://wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/?lang=uk/",
		"https://wikipedia.ua/",
		"https://ua.wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/",
		"https://uk.wikipedia.org/wiki/%d0%93%d0%be%d0%bb%d0%be%d0%b2%d0%bd%d0%b0_%d1%81%d1%82%d0%be%d1%80%d1%96%d0%bd%d0%ba%d0%b0/",
	}

	if !EqualArrays(got, expected) {
		t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
			strings.Join(expected, ", "))
	}
}

func TestCheckLang(t *testing.T) {
	testLinks := []string {"https://ru.wikipedia.org/wiki/%D0%9C%D0%BE%D1%81%D0%BA%D0%B2%D0%B0", "https://gordonua.com/"}
	var pageLangs []string
	var pageLang string
	for _, mainLink := range testLinks {
		allUkrLinks := getAllSiteUkrLinks(mainLink)
		pageLang = "not ukrainian"

		collector := colly.NewCollector()
		site := Site{}
		hyperlinksSet := NewSet()
		var mum map[string][]string
		mum = make(map[string][]string)

		setUpCollector(collector, &site, &mum, hyperlinksSet)
		for _, newLink := range allUkrLinks {
			collector.Visit(newLink)
			fmt.Println("site.Title", site.Title)

			site.Content = strings.TrimSpace(strings.Join(mum["p"], " \n ") +
				strings.Join(mum["li"], " \n ") + strings.Join(mum["div"], " \n ") +
				strings.Join(mum["article"], " \n "))

			if checkLang(&site.Content, &site.Title, "Ukrainian") {
				mainLink = newLink
				pageLang = "ukrainian"
				fmt.Println("checkMainPageLang() Language is Ukrainian ", mainLink)
				break
			}
		}

		pageLangs = append(pageLangs, pageLang)
	}

	for i := 0; i < len(testLinks); i++ {
		fmt.Println(testLinks[i], " -- ", pageLangs[i])
	}

	//if !EqualArrays(got, expected) {
	//	t.Errorf("got -- %s; expected -- %s", strings.Join(got, ", "),
	//		strings.Join(expected, ", "))
	//}
}
