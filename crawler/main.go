package main

import (
		"fmt"
		"encoding/json"
		"time"
		"github.com/gocolly/colly"
		"io/ioutil"
		"bufio"
		"log"
		"os"
		"strings"
		)

type Site struct {
	Title   string    `json:"title"`
	Link    string    `json:"link"`
	Content string `json:"inline_text"`
	Hyperlinks    []string  `json:hyperlinks`
	AddedAt time.Time `json:"added_at_time"`
}

func writeJSON(data []Site, writefile string){
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil{
		log.Println("error while writing a file")
		return
	}

	_ = ioutil.WriteFile(writefile, file, 0644)
}

func readLines(path string) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}
 func crawl(link string,  jobs <-chan int, results chan<- int, lst *[]Site){
	allSites := make([]Site, 0)
	 
	 // colector -> colly scrape interface
	 collector := colly.NewCollector()

	 
	 // happens on the request
	 collector.OnRequest(func (request *colly.Request){
		 fmt.Println("Visiting", request.URL.String())
		})
		
	collector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		results <- 0 
		return 
	})

	var mum map[string][]string
	mum = make(map[string][]string)
	// var index int = 0;
	collector.OnHTML("p", func(element *colly.HTMLElement){
		mum[element.Name] = append(mum[element.Name], element.Text)
		site := Site{ 
					  Title: "",
					  Link: (element.Request).URL.String(),
					  Hyperlinks: make([]string,0),
					  }

		allSites = append(allSites, site)
		
	})

	
	collector.OnHTML("title", func(element *colly.HTMLElement){
		allSites[0].Title = element.Text
	})
	

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link != "" {
			allSites[0].Hyperlinks = append(allSites[0].Hyperlinks, link)
			}
	 })

	collector.OnError(func(response *colly.Response, err error) {
		results <- -1
	})
	collector.Visit(link)
	// fmt.Println(mum["p"])
	// var l []string = *a
	allSites[0].Content = strings.Join(mum["p"], " \n ")
	*lst = append(*lst, allSites[0])
	results <- 0 
}

func main() {
	links, err := readLines("links.txt")
    if err != nil {
        log.Fatalf("readLines: %s", err)
    }
    var numJobs = len(links)
    jobs := make(chan int, numJobs)
    results := make(chan int, numJobs)

	var sites []Site
	sites = make([]Site, 0)	
    for w := 1; w <= numJobs; w++ {
		go crawl(links[w-1], jobs, results,&sites)
    }

    for j := 1; j <= numJobs; j++ {
        jobs <- j
    }
    close(jobs)

    for a := 1; a <= numJobs; a++ {
        <-results
    }
	writeJSON(sites, "out.json")
}
