package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type responseLinks struct {
	Links []string `json:"links"`
}

func writeJSON(data <-chan Site, writefile string) {
	allSites := make([]Site, 0)
	for msg := range data {
		allSites = append(allSites, msg)
	}

	file, err := json.MarshalIndent(allSites, "", " ")
	if err != nil {
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

func crawl(lst chan<- Site, queue chan string, done, ks chan bool, wg *sync.WaitGroup) {
	for true {
		select {
		case k := <-queue:
			allSites := make([]Site, 0)

			collector := colly.NewCollector()

			collector.OnRequest(func(request *colly.Request) {
				fmt.Println("Visiting", request.URL.String())
			})

			collector.OnResponse(func(response *colly.Response) {
				if response.StatusCode != 200 {
					fmt.Println(response.StatusCode)
				}
			})

			collector.OnError(func(r *colly.Response, err error) {
				fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
			})

			var mum map[string][]string
			mum = make(map[string][]string)
			site := Site{
				Title:      "",
				Link:       "",
				Hyperlinks: make([]string, 0),
			}

			collector.OnHTML("p", func(element *colly.HTMLElement) {
				mum[element.Name] = append(mum[element.Name], element.Text)
				site.Link = (element.Request).URL.String()
			})

			collector.OnHTML("li", func(element *colly.HTMLElement) {
				mum[element.Name] = append(mum[element.Name], element.Text)
				site.Link = (element.Request).URL.String()
			})

			collector.OnHTML("article", func(element *colly.HTMLElement) {
				mum[element.Name] = append(mum[element.Name], element.Text)
				site.Link = (element.Request).URL.String()
			})

			collector.OnHTML("html", func(e *colly.HTMLElement) {
				site.Title = e.ChildAttr(`meta[property="og:title"]`, "content")
			})

			collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
				link := e.Request.AbsoluteURL(e.Attr("href"))
				if link != "" {
					site.Hyperlinks = append(site.Hyperlinks, link)
				}
			})

			collector.Visit(k)

			allSites = append(allSites, site)

			allSites[0].Content = strings.Join(mum["p"], " \n ") + strings.Join(mum["li"], " \n ") + strings.Join(mum["article"], " \n ")
			lst <- allSites[0]
			defer wg.Done()
			done <- true
		case <-ks:
			return
		}
	}

}

func main() {
	// Perform health-check
	for {
		_, err_elastic := http.Get(os.Getenv("ELASTICSEARCH_URL"))
		_, err_manager := http.Get(os.Getenv("TASK_MANAGER_URL") + "/health_check")
		fmt.Println("Waiting for Elasticsearch and Task Manager to be alive.")
		if err_elastic == nil && err_manager == nil {
			break
		}
		time.Sleep(time.Second)
	}
	// Elasticsearch and Task Manager have started. The program begins

	// ------ get links from TaskManager
	resp, err := http.Get(os.Getenv("TASK_MANAGER_URL") + "/task_manager/api/v1.0/get_links")
	//resp, err := http.Get("https://jsonplaceholder.typicode.com/posts/1")

	// check for response error
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(body))

	res := responseLinks{}
	json.Unmarshal(body, &res)

	links := res.Links
	fmt.Println("response Links  -- ", links)
	// ------ end getting links from TaskManager

	//fmt.Println("res  ", res)

	//if false {
	//links, err := readLines("links.txt")
	//if err != nil {
	//	log.Fatalf("readLines: %s", err)
	//}

	// elastic connection
	esClient := elasticConnect()

	insertIdxName := os.Getenv("INDEX_ELASTIC_COLLECTED_DATA")
	titleStr := "start index"
	contentStr := "first content1"
	setIndexFirstId(esClient, insertIdxName, titleStr, contentStr)

	indexLastId := indexGetLastId(esClient, insertIdxName)

	indexLastIdInt, err := strconv.Atoi(indexLastId)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	indexLastIdInt += 1
	fmt.Println("my indexLastId", indexLastIdInt)
	// end elastic connection

	//if false {
	var numberOfJobs = len(links)
	var wg sync.WaitGroup

	sites := make(chan Site, 50)
	killsignal := make(chan bool)
	q := make(chan string)
	done := make(chan bool)

	numberOfWorkers := 2
	for i := 0; i < numberOfWorkers; i++ {
		go crawl(sites, q, done, killsignal, &wg)
	}

	for j := 0; j < numberOfJobs; j++ {
		go func(j int) {
			wg.Add(1)
			q <- links[j]
		}(j)
	}

	for c := 0; c < numberOfJobs; c++ {
		<-done
	}

	close(killsignal)
	wg.Wait()

	close(sites)

	//time.Sleep(1)
	//crawl(links[numJobs-1], jobs, results, &sites)
	//writeJSON(sites, "out.json")

	// write to elastic
	allSites := make([]Site, 0)
	for msg := range sites {
		allSites = append(allSites, msg)
	}
	elasticInsert(esClient, allSites, &insertIdxName, &indexLastIdInt)
	//}
}
