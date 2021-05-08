package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync/atomic"

	//"time"

	//"strconv"
	"strings"
	"sync"
	//"time"
)

func writeSliceJSON(data []Site, writefile string) {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println("error while writing a file")
		return
	}

	_ = ioutil.WriteFile(writefile, file, 0644)
}

func writeChannelJSON(data <-chan Site, writefile string) {

	allSites := make([]Site, 0)
	for msg := range data {
		// if msg.Title == ""
		{
			allSites = append(allSites, msg)
		}
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

func visitLink(lst chan<- Site, link string, visited *SafeSetOfLinks, id int) (failed error) {
	fmt.Println("parsing link -- ", link)

	var site Site
	hyperlinksSet := NewSet()
	//collector := colly.NewCollector(
	//	colly.AllowedDomains("https://organexpressions.com","organexpressions.com", "https://www.organexpressions.com", "www.organexpressions.com",
	//						 "https://oneessencehealing.com","oneessencehealing.com", "https://www.oneessencehealing.com", "www.oneessencehealing.com"),
	//)

	collector := colly.NewCollector()

	collector.OnRequest(func(request *colly.Request) {
		// if (request.URL.String() =="https://oneessencehealing.com/"){
		fmt.Println("Visiting", request.URL.String())
		// }
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
		site.Link = strings.TrimSpace((element.Request).URL.String())
	})

	collector.OnHTML("li", func(element *colly.HTMLElement) {
		mum[element.Name] = append(mum[element.Name], element.Text)
		site.Link = strings.TrimSpace((element.Request).URL.String())
	})

	collector.OnHTML("article", func(element *colly.HTMLElement) {
		mum[element.Name] = append(mum[element.Name], element.Text)
		site.Link = strings.TrimSpace((element.Request).URL.String())
	})

	collector.OnHTML("head", func(element *colly.HTMLElement) {
		link := strings.TrimSpace(element.Attr("title"))
		site.Title = site.Title + " " + link
		site.Link = strings.TrimSpace((element.Request).URL.String())
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
		site.Title = site.Title + " " + e.ChildAttr(`meta[property="og:title"]`, "content") + " " +
			e.ChildText("title") + e.DOM.Find("title").Text()
	})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := strings.TrimSpace(e.Request.AbsoluteURL(e.Attr("href")))
		if link != "" && len(link) > 5 && link[:5] == "https" {
			// clear from link parameters
			startLinkParameters := strings.Index(link, "?")

			if startLinkParameters > 0 {
				link = link[:startLinkParameters]
			}

			hyperlinksSet.Add(link)
		}

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

	found := visited.checkIfContains(link)

	if !found {
		// *visited = append(*visited, link)
		visited.addLink(link)
		collector.Visit(link)
	} else {
		fmt.Println("checkIfContains ", link, "visited -- ", found)
		return
	}

	//site.Hyperlinks = hyperlinksSet.dict
	site.Hyperlinks = make([]string, 0)
	for link := range hyperlinksSet.dict {
		site.Hyperlinks = append(site.Hyperlinks, link)
	}

	site.Content = strings.TrimSpace(strings.Join(mum["p"], " \n ") +
		strings.Join(mum["li"], " \n ") + strings.Join(mum["article"], " \n "))
	// _, found := Find(*visited, link)
	// if !found {
	// 	*visited = append(*visited, link)
	// 	collector.Visit(link)
	// }

	if site.Link == "" {
		return
	}

L:
	for true {
		select {
		case lst <- site:
			break L
		default:

		}
	}

	//fmt.Println("visited -- ", visited)
	return

	//for _, s := range site.Hyperlinks {
	//	visitLink(lst, s, visited, id)
	//	// TODO - delete
	//	//break
	//}
	//
	//if os.Getenv("DEBUG") == "true" {
	//	fmt.Println("after cycle")
	//}
	//
	//return
}

func crawl(lst chan<- Site, linksQueue chan string, done, ks chan bool,
	wg *sync.WaitGroup, visited *SafeSetOfLinks, failedLinks chan map[string]string, id int) {

	for true {
		select {
		case link := <-linksQueue:
			// site Side
			var failed error
			failed = visitLink(lst, link, visited, id)

			if failed == nil {

			}

			done <- true
			if failed != nil {
				m := make(map[string]string)
				m["link"] = link
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

	// load .env file
	err := godotenv.Load(path.Join("..", "crawlers-env.env"))
	//err := godotenv.Load(path.Join("crawlers-env.env"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// ------ get links from TaskManager
	resp, err := http.Get(os.Getenv("TASK_MANAGER_URL") + "/task_manager/api/v1.0/get_links")

	// check for response error
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	res := responseLinks{}
	json.Unmarshal(body, &res)

	// ================== setup configuration ==================
	// TODO: change after testing
	//links := append(res.Links[:20], "https://www.google.com/")
	links := res.Links[:45]
	fmt.Println("start len(links) -- ", len(links))
	numberOfWorkers := 2
	var numberOfJobs = len(links)

	if os.Getenv("DEBUG") == "true" {
		fmt.Println("response Links  -- ", links)
	}

	// ------ end getting links from TaskManager

	// elastic connection
	esClient := elasticConnect()

	insertIdxName := os.Getenv("INDEX_ELASTIC_COLLECTED_DATA")
	titleStr := "start index"
	contentStr := "first content1"
	setIndexFirstId(esClient, insertIdxName, titleStr, contentStr)

	indexLastIdInt := indexGetLastId(esClient, insertIdxName)

	indexLastIdInt += 1
	fmt.Println("my indexLastId", indexLastIdInt)
	// end elastic connection

	// used for testing
	//if false {
	var wg sync.WaitGroup

	sliceSites := SafeListOfSites{actualSites: make([]Site, 0)}
	sites := make(chan Site, len(links)+1)
	failedLinks := make(chan map[string]string, len(links)+1)
	killSignal := make(chan bool)
	finish := make(chan bool)

	allParsedLinks := newSafeSetOfLinks()

	var numAddedSites uint64
	go func(finish chan bool, sliceSites *SafeListOfSites, sites <-chan Site,
		allParsedLinks *SafeSetOfLinks, numAddedSites *uint64) {
		wg.Add(1)

	F:
		for true {
			select {
			// take from channel of parsed sites and insert in elasticsearch
			case site := <-sites:
				//_, found := sliceSites.checkIfContains(site)
				//if !found {
				//	sliceSites.addSite(site)
				//}

				found := allParsedLinks.checkIfContains(site.Link)
				fmt.Println("found in allParsedLinks -- ", found)
				if !found {
					fmt.Println("add Link -- ", site.Link)
					allParsedLinks.addLink(site.Link)
					sliceSites.addSite(site)
					atomic.AddUint64(numAddedSites, 1)
				}

				if len(sliceSites.actualSites) >= 5 {
					elasticInsert(esClient, &sliceSites.actualSites, &insertIdxName, &indexLastIdInt)
				}

			case <-finish:
				if len(sites) == 0 {
					break F
				}
				//case <-killSignal:
				//	// TODO:
				//	numFinishedRoutins++
				//	if numFinishedRoutins == numberOfWorkers {
				//		break F
				//	}

			}
		}

		fmt.Println("len(sites) -- ", len(sites))

		if len(sliceSites.actualSites) != 0 {
			elasticInsert(esClient, &sliceSites.actualSites, &insertIdxName, &indexLastIdInt)
		}

		defer wg.Done()
	}(finish, &sliceSites, sites, allParsedLinks, &numAddedSites)

	linksQueue := make(chan string)
	done := make(chan bool)

	// TODO: replace visited variable in crawl function
	visited := newSafeSetOfLinks()
	for i := 0; i < numberOfWorkers; i++ {
		go crawl(sites, linksQueue, done, killSignal, &wg, visited, failedLinks, i)
	}

	for j := 0; j < numberOfJobs; j++ {
		// TODO: check duplicate at the beginning when take domain
		go func(j int) {
			wg.Add(1)

			// avoid http links and complete to a full link of domain
			if !strings.Contains(links[j], "http") {
				linksQueue <- "https://" + links[j]
			} else {
				linksQueue <- links[j]
			}

		}(j)
	}

	for c := 0; c < numberOfJobs; c++ {
		<-done
	}
	fmt.Println("len(done) -- ", len(done))

	close(killSignal)
	close(finish)
	wg.Wait()

	close(sites)
	close(failedLinks)

	fmt.Println("numAddedSites -- ", numAddedSites)

	//time.Sleep(1)
	//crawl(links[numJobs-1], jobs, results, &sites)
	//writeJSON(sites, "out.json")
}
