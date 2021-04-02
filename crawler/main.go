package main

import (
	//"bufio"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	//"time"
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

func visit_link(link string) (site Site, failed error) {
	collector := colly.NewCollector()

	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting", request.URL.String())
	})

	collector.OnResponse(func(response *colly.Response) {
		if response.StatusCode != 200 {
			fmt.Println(response.StatusCode)
		}
	})

	collector.OnError(func(response *colly.Response, err error) {
		failed = err
		if response.StatusCode != 200 && response.StatusCode != 0 {
			fmt.Println(response.StatusCode)
		}
	})

	var mum map[string][]string
	mum = make(map[string][]string)

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

	collector.OnHTML("head", func(element *colly.HTMLElement) {
		link := element.Attr("title")
		site.Title = site.Title + " " + link
		site.Link = (element.Request).URL.String()
		// site.Title = site.Title + " " + element.ChildAttr(`title`,)
		site.Title = site.Title + " " + element.ChildText("title") + " " + element.DOM.Find("title").Text()
	})

	collector.OnHTML("title", func(element *colly.HTMLElement) {
		site.Title = site.Title + " " + element.Text
	})

	collector.OnHTML("h1", func(element *colly.HTMLElement) {
		site.Title = site.Title + " " + element.Text
	})

	collector.OnHTML("html", func(e *colly.HTMLElement) {
		site.Title = site.Title + " " + e.ChildAttr(`meta[property="og:title"]`, "content") + " " + e.ChildText("title") + e.DOM.Find("title").Text()
	})

	// c.OnHTML("html", func(e *colly.HTMLElement) {
	// 	if strings.EqualFold(e.ChildAttr(`meta[property="og:type"]`, "content"), "article") {
	// 		// Find the emoji page title
	// 		fmt.Println("Emoji: ", e.ChildText("article h1"))
	// 		// Grab all the text from the emoji's description
	// 		fmt.Println("Description: ", e.ChildText("article .description p"))
	// 	}
	// })
	// 	site.Title

	// if strings.EqualFold(e.ChildAttr(`meta[property="og:title"]`, "content"), "article") {
	// 	// Find the emoji page title
	// 	fmt.Println("Emoji: ", e.ChildText("article h1"))
	// 	// Grab all the text from the emoji's description
	// 	fmt.Println("Description: ", e.ChildText("article .description p"))
	// }
	// })

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link != "" {
			site.Hyperlinks = append(site.Hyperlinks, link)
		}
	})

	collector.Visit(link)

	site.Content = strings.Join(mum["p"], " \n ") + strings.Join(mum["li"], " \n ") + strings.Join(mum["article"], " \n ")

	return
}

func crawl(lst chan<- Site, queue chan string, done, ks chan bool, wg *sync.WaitGroup, failedLinks chan map[string]string) {
	for true {
		select {
		case k := <-queue:
			// site Side
			var site Site
			var failed error
			site, failed = visit_link(k)

			if failed == nil {
				lst <- site
			}

			done <- true
			if failed != nil {
				// fmt.Println()
				m := make(map[string]string)
				m["link"] = k
				m["error"] = failed.Error()
				failedLinks <- m
			}
			defer wg.Done()
		case <-ks:
			return
		}
	}

}

func main() {
	// Perform health-check

	//for {
	//	_, err_elastic := http.Get(os.Getenv("ELASTICSEARCH_URL"))
	//	_, err_manager := http.Get(os.Getenv("TASK_MANAGER_URL") + "/health_check")
	//	fmt.Println("Waiting for Elasticsearch and Task Manager to be alive.")
	//	if err_elastic == nil && err_manager == nil {
	//		break
	//	}
	//	time.Sleep(time.Second)
	//}

	// Elasticsearch and Task Manager have started. The program begins

	//log.Println("log.Println " + os.Getenv("TASK_MANAGER_URL"))
	//log.Println(os.Environ())

	// load .env file
	err := godotenv.Load(path.Join("..", "crawlers-env.env"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// ------ get links from TaskManager
	resp, err := http.Get(os.Getenv("TASK_MANAGER_URL") + "/task_manager/api/v1.0/get_links")
	//resp, err := http.Get("http://localhost:5000" + "/task_manager/api/v1.0/get_links")
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

	if os.Getenv("DEBUG") == "true" {
		fmt.Println("response Links  -- ", links)
	}

	// ------ end getting links from TaskManager

	//fmt.Println("res  ", res)

	//if false {
	//links, err := readLines("links.txt")
	//if err != nil {
	//	log.Fatalf("readLines: %s", err)
	//}

	// elastic connection
	esClient := elasticConnect()

	// TODO: uncomment
	insertIdxName := os.Getenv("INDEX_ELASTIC_COLLECTED_DATA")
	//insertIdxName := "t_english_sites32"
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

	sites := make(chan Site, len(links)+1)
	failedLinks := make(chan map[string]string, len(links)+1)

	killsignal := make(chan bool)

	queue := make(chan string)

	done := make(chan bool)

	numberOfWorkers := 20
	for i := 0; i < numberOfWorkers; i++ {
		go crawl(sites, queue, done, killsignal, &wg, failedLinks)
	}

	for j := 0; j < numberOfJobs; j++ {

		// select {
		// case k:= queue<-
		// }
		go func(j int) {
			wg.Add(1)
			if !strings.Contains(links[j], "https://") {
				queue <- "https://" + links[j]
			} else {
				queue <- links[j]
			}

		}(j)
	}

	for c := 0; c < numberOfJobs; c++ {
		<-done
		// <-failedLinks
	}

	close(killsignal)

	wg.Wait()
	// close(failedLinks)

	close(sites)

	close(failedLinks)

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
