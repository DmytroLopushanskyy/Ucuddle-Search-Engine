package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync/atomic"
	"time"
	//"strconv"
	"strings"
	"sync"
	//"time"
)

func writeSliceJSON(data []Site, writefile string) {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		standardLogger.Println("error while writing a file")
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
		standardLogger.Println("error while writing a file")
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

func visitLink(lst chan<- Site, mainLink string,
	visited *SafeSetOfLinks, id int, numInternalPage *uint64, domain string) (failed error) {

	var MaxInternalPages uint64
	MaxInternalPages, _ = strconv.ParseUint(os.Getenv("MAX_LIMIT_INTERNAL_PAGES"), 10, 64)
	if *numInternalPage >= MaxInternalPages {
		return
	}

	var site Site
	hyperlinksSet := NewSet()


	if !strings.Contains(link, domain){
		return 
	}

	collector := colly.NewCollector(
							
	)
	collector.OnRequest(func(request *colly.Request) {
		standardLogger.Println("Visiting", request.URL.String())
	})

	collector.OnResponse(func(response *colly.Response) {
		if response.StatusCode != 200 {
			standardLogger.Println(response.StatusCode)
		}
	})

	collector.OnError(func(response *colly.Response, err error) {
		failed = err
		if response.StatusCode != 200 && response.StatusCode != 0 {
			standardLogger.Println(response.StatusCode)
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
		// site.Title = site.Title + " " + link
		site.Link = strings.TrimSpace((element.Request).URL.String())
		// site.Title = site.Title + " " + element.ChildAttr(`title`,)
		
		if site.Title == " "{
			site.Title = element.ChildText("title") 
		}

		if site.Title == " "{
			site.Title = element.DOM.Find("title").Text()
		}

	})

	collector.OnHTML("title", func(element *colly.HTMLElement) {
		if site.Title == " "{
			site.Title = element.Text
		}
	})

	collector.OnHTML("h1", func(element *colly.HTMLElement) {
		if site.Title == " "{
			site.Title = element.Text
		}
	})

	collector.OnHTML("html", func(e *colly.HTMLElement) {
		if site.Title == " "{
			e.ChildAttr(`meta[property="og:title"]`, "content")
		}

		if site.Title == " "{
			e.ChildText("title")
		}

		if site.Title == " "{
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

			if link[len(link)- 1: len(link)] == "/" {
				link = link[:len(link) - 1]
			}

			hyperlinksSet.Add(link)
		}

	})

	found := visited.checkIfContains(mainLink)

	if !found {
		visited.addLink(mainLink)
		collector.Visit(mainLink)
		atomic.AddUint64(numInternalPage, 1)
	} else {
		standardLogger.Println("checkIfContains ", mainLink, "visited -- ", found)
		return
	}

	site.Hyperlinks = make([]string, 0)
	for link := range hyperlinksSet.dict {
		site.Hyperlinks = append(site.Hyperlinks, link)
	}

	site.Content = strings.TrimSpace(strings.Join(mum["p"], " \n ") +
		strings.Join(mum["li"], " \n ") + strings.Join(mum["article"], " \n "))


	if site.Link == "" {
		return
	}

	if site.Link[len(site.Link)- 1: len(site.Link)] == "/" {
		site.Link = site.Link[:len(site.Link) - 1]
	}

L:
	for true {
		select {
		case lst <- site:
			break L
		default:

		}
	}

	for _, s := range site.Hyperlinks {
		visitLink(lst, s, visited, id, numInternalPage, domain)
	}

	return
}

func crawl(lst chan<- Site, linksQueue chan [2]string, done, ks chan bool,
	wg *sync.WaitGroup, visited *SafeSetOfLinks, failedLinks chan map[string]string, id int) {

	for true {
		select {
		case link := <-linksQueue:
			// site Side
			var failed error
			var numInternalPage uint64
			numInternalPage = 0
			failed = visitLink(lst, link[0], visited, id, &numInternalPage, link[0])

			if failed == nil {

			}

			done <- true
			if failed != nil {
				m := make(map[string]string)
				m["link"] = link[0]
				m["error"] = failed.Error()
				failedLinks <- m
			}

			// TODO: maybe add in if failed == nil
			setParsedLink(link[1])

			defer wg.Done()
		case <-ks:
			return
		}
	}
}

func crawlLinksPackage(esClient *elasticsearch.Client, insertIdxName string, links *[][2]string,
	numberOfWorkers int, numberOfJobs int, lenLinks int) uint64 {

	var wg sync.WaitGroup

	sliceSites := SafeListOfSites{actualSites: make([]Site, 0)}
	sites := make(chan Site, lenLinks+1)
	failedLinks := make(chan map[string]string, lenLinks+1)
	killSignal := make(chan bool)
	finishElasticInsert := make(chan bool)

	allParsedLinks := newSafeSetOfLinks()

	var numAddedPages uint64
	go func(finishElasticInsert chan bool, sliceSites *SafeListOfSites, sites <-chan Site,
		allParsedLinks *SafeSetOfLinks, numAddedPages *uint64) {
		wg.Add(1)

	F:
		for true {
			select {
			// take from channel of parsed sites and insert in elasticsearch
			case site := <-sites:
				found := allParsedLinks.checkIfContains(site.Link)
				if !found {
					allParsedLinks.addLink(site.Link)
					sliceSites.addSite(site)
					atomic.AddUint64(numAddedPages, 1)
				}

				packageSize, _ := strconv.Atoi(os.Getenv("NUM_SITES_IN_PACKAGE"))
				if len(sliceSites.actualSites) >= packageSize {
					elasticInsert(esClient, &sliceSites.actualSites, &insertIdxName, 0)
				}

			case <-finishElasticInsert:
				if len(sites) == 0 {
					break F
				}
			}
		}

		standardLogger.Println("len(sites) -- ", len(sites))

		if len(sliceSites.actualSites) != 0 {
			elasticInsert(esClient, &sliceSites.actualSites, &insertIdxName, 0)
		}

		defer wg.Done()
	} (finishElasticInsert, &sliceSites, sites, allParsedLinks, &numAddedPages)

	linksQueue := make(chan [2]string)
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
			if !strings.Contains((*links)[j][0], "http") {
				linksQueue <- [2]string {"https://" + (*links)[j][0], (*links)[j][1]}
			} else {
				linksQueue <- (*links)[j]
			}

		}(j)
	}

	for c := 0; c < numberOfJobs; c++ {
		<-done
	}
	standardLogger.Println("at the end of code len(done) -- ", len(done))

	close(killSignal)
	close(finishElasticInsert)
	wg.Wait()

	close(sites)
	close(failedLinks)

	return numAddedPages
}

func main() {
	os.Setenv("COLLY_IGNORE_ROBOTSTXT", "n")

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
	startCode := time.Now()

	// load .env file
	err := godotenv.Load(path.Join("shared_vars.env"))
	//err := godotenv.Load(path.Join("..", "shared_vars.env"))

	if err != nil {
		standardLogger.Fatal("Error loading .env file")
	}

	// ================== setup configuration ==================
	numberOfWorkers := 8

	// ================== elastic connection ==================
	esClient := elasticConnect()

	insertIdxName := os.Getenv("INDEX_ELASTIC_COLLECTED_DATA")
	titleStr := "start index"
	contentStr := "first content1"
	setIndexFirstId(esClient, insertIdxName, titleStr, contentStr)
	// end elastic connection

	// ================== set last site id in Task Manager ==================
	postBody, _ := json.Marshal(map[string]string {
		"1":  "1",
	})
	responseBody := bytes.NewBuffer(postBody)
	http.Post(os.Getenv("TASK_MANAGER_URL") + os.Getenv("TASK_MANAGER_ENDPOINT_SET_LAST_SITE_ID"),
		"application/json", responseBody)


	var totalNumAddedPages uint64
	var continueFlag bool
	iteration := 0
	indexesElasticLinks := strings.Split(os.Getenv("INDEXES_ELASTIC_LINKS"), " ")

	for j := 0; j < len(indexesElasticLinks); j++ {
		endedIdxLinksCounter := 0

		for true {
			iteration++

			standardLogger.Println("Start getDomainsToParse global iteration ", iteration)

			// ================== get links from TaskManager ==================
			res := responseLinks{}

			if endedIdxLinksCounter == 0 {
				getDomainsToParse(&res, false)
				standardLogger.Println("get domains taken: false, parsed: false")
			} else if endedIdxLinksCounter == 1 {
				getDomainsToParse(&res, true)
				standardLogger.Println("get domains taken: true, parsed: false")
			} else {
				standardLogger.Println("reached end of the current index_name, global iteration over indexes_names -- ", j)
				break
			}

			continueFlag = false

			Block {
				Try: func() {
					if res.Links[0][0] == "links ended" {
						endedIdxLinksCounter++
						continueFlag = true
					}
				},
				Catch: func(e Exception) {
					standardLogger.Warn("Caught %v\n", e)
					continueFlag = true
				},
			}.Do()

			if continueFlag {
				continue
			}

			links := res.Links
			standardLogger.Println("start len(links) -- ", len(links))
			standardLogger.Println("first taken link -- ", links[0])
			standardLogger.Println("last taken link -- ", links[len(links)-1])
			var numberOfJobs = len(links)

			// ------ end getting links from TaskManager

			totalNumAddedPages += crawlLinksPackage(esClient, insertIdxName, &links,
				numberOfWorkers, numberOfJobs, len(links))

			standardLogger.Println("Iteration  ", iteration, ", after this iteration totalNumAddedPages -- ", totalNumAddedPages)
			finishedCode := time.Now()

			standardLogger.Println("Iteration  ", iteration,
				", total work time after this iteration -- ", finishedCode.Sub(startCode), "\n\n")
		}
	}

	//time.Sleep(1)
	//crawl(links[numJobs-1], jobs, results, &sites)
	//writeJSON(sites, "out.json")
}
