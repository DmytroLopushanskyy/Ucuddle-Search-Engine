package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"path"

	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

const PROXY_PERCENT_LOWER_LIMIT float64 = 0.5
var DEAD_WORKERS int32 = 0

var CUR_INDEX_LINKS_ID uint
var NUM_INDEXES_LINKS_ID uint
var NUM_LINKS_PER_CRAWLER float64


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
	visited *SafeSetOfLinks, id int, numInternalPage *uint64,
	domain string) (failed error) {

	var MaxInternalPages uint64
	MaxInternalPages, _ = strconv.ParseUint(os.Getenv("MAX_LIMIT_INTERNAL_PAGES"), 10, 64)
	if *numInternalPage > MaxInternalPages {
		return
	}

	var site Site
	hyperlinksSet := NewSet()

	var pageLang string
	var mum map[string][]string
	mum = make(map[string][]string)

	if !strings.Contains(mainLink, domain) {
		return
	}

	collector := colly.NewCollector()

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

	collector.OnHTML("p", func(element *colly.HTMLElement) {
		if len(element.Text) > 1 {
			content := strings.Join(strings.Fields(element.Text), " ")
			mum[element.Name] = append(mum[element.Name], content)
			site.Link = strings.TrimSpace((element.Request).URL.String())
		}
	})

	collector.OnHTML("div", func(element *colly.HTMLElement) {
		if len(element.Text) > 1 {
			content := strings.Join(strings.Fields(element.Text), " ")
			mum[element.Name] = append(mum[element.Name], content)
			site.Link = strings.TrimSpace((element.Request).URL.String())
		}
	})

	collector.OnHTML("li", func(element *colly.HTMLElement) {
		if len(element.Text) > 1 {
			content := strings.Join(strings.Fields(element.Text), " ")
			mum[element.Name] = append(mum[element.Name], content)
			site.Link = strings.TrimSpace((element.Request).URL.String())
		}
	})

	collector.OnHTML("article", func(element *colly.HTMLElement) {
		if len(element.Text) > 1 {
			content := strings.Join(strings.Fields(element.Text), " ")
			mum[element.Name] = append(mum[element.Name], content)
			site.Link = strings.TrimSpace((element.Request).URL.String())
		}
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

	found := visited.checkIfContains(&mainLink)

	if found {
		standardLogger.Println("checkIfContains ", mainLink, "visited -- ", found)
		return
	}

	lenLink := len(domain)
	if domain[lenLink-1:lenLink] != "/" {
		domain += "/"
	}

	if strings.Contains(domain, "uk-ua.") ||
		strings.Contains(domain, "?lang=uk") || strings.Contains(domain, "/uk/") {
		pageLang = "ukr"
	} else if strings.Contains(domain, "ru-ru.") || strings.Contains(domain, "https://ru.") ||
		strings.Contains(domain, "?lang=ru") || strings.Contains(domain, "/ru/") {
		pageLang = "ru"
	}

	visited.addLink(&mainLink)
	collector.Visit(mainLink)

	// TODO: reminder that numInternalPage != num pages saved in database
	atomic.AddUint64(numInternalPage, 1)

	site.Content = strings.TrimSpace(strings.Join(mum["p"], "\n") +
		strings.Join(mum["li"], "\n") + strings.Join(mum["div"], "\n") +
		strings.Join(mum["article"], "\n"))

	if pageLang != "ukr" && pageLang != "ru" {
		siteLang := checkLang(&site.Content, &site.Title)
		if siteLang == "Ukrainian" {
			pageLang = "ukr"
		} else if siteLang == "Russian" {
			pageLang = "ru"
		}
	}

	if pageLang != "ukr" && pageLang != "ru" {
		standardLogger.Println("Site does not have ukr and ru translations ", mainLink)
		return
	}

	standardLogger.Println("Website supports needed languages -- ", pageLang, " ", mainLink)
	site.Lang = pageLang

	site.Hyperlinks = make([]string, 0)
	for link := range hyperlinksSet.dict {
		site.Hyperlinks = append(site.Hyperlinks, link)
	}

	if site.Link == "" {
		return
	}

	if site.Link[len(site.Link)-1:len(site.Link)] == "/" {
		site.Link = site.Link[:len(site.Link)-1]
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
	wg *sync.WaitGroup, failedLinks chan map[string]string, id int,
	proxyIsFree *int32, numVisitedDomains *uint32, finishElasticInsert chan bool) {

	var domain string
	var endDomainPos int
	for true {
		select {
		case link := <-linksQueue:
			// site Side
			var failed error
			var numInternalPage uint64
			numInternalPage = 0
			domain = link[0]

			endDomainPos = findNthSymbol(&link[0], "/", 3)
			if endDomainPos != -1 {
				domain = domain[:endDomainPos]
			}

			visited := newSafeSetOfLinks()
			failed = visitLink(lst, link[0], visited, id, &numInternalPage, domain)
			atomic.AddUint32(numVisitedDomains, 1)

			standardLogger.Info("total numVisitedDomains -- ", *numVisitedDomains)

			if failed == nil {

			}

			//done <- true
			if failed != nil {
				m := make(map[string]string)
				m["link"] = link[0]
				m["error"] = failed.Error()
				failedLinks <- m
			}

			setParsedLink(link[1])

			standardLogger.Debug("len(linksQueue) -- ", len(linksQueue),
				"\n int(PROXY_PERCENT_LOWER_LIMIT * NUM_LINKS_PER_CRAWLER) + 1 -- ",
				int(PROXY_PERCENT_LOWER_LIMIT * NUM_LINKS_PER_CRAWLER) + 1,
				"\n proxyIsFree -- ", *proxyIsFree)
			if len(linksQueue) < (int(PROXY_PERCENT_LOWER_LIMIT * NUM_LINKS_PER_CRAWLER) + 1) && *proxyIsFree == 1 {
				atomic.AddInt32(proxyIsFree, -1)
				standardLogger.Info("achieved limit of num domains in linksQueue, proxyIsFree -- ", *proxyIsFree)
				proxyLoadNewDomains(linksQueue, proxyIsFree, wg, &CUR_INDEX_LINKS_ID)
				standardLogger.Info("after proxyLoadNewDomains proxyIsFree -- ", *proxyIsFree)
				standardLogger.Info("after proxyLoadNewDomains len(linksQueue) -- ", len(linksQueue))
			}

			// TODO: actually it does not stop all crawlers
			//  which are running on different machines. So it should be fixed
			// achieved "links ended" in each of indexes of links in elasticsearch
			if CUR_INDEX_LINKS_ID >= NUM_INDEXES_LINKS_ID {
				standardLogger.Info("linksQueue is empty -- return")
				//finishElasticInsert <- true
				//ks <- true
				atomic.AddInt32(&DEAD_WORKERS, 1)

				return
			}

			//defer wg.Done()
		//case <-ks:
		default:
			if CUR_INDEX_LINKS_ID >= NUM_INDEXES_LINKS_ID {
				//finishElasticInsert <- true
				standardLogger.Info("Got killsignal")
				atomic.AddInt32(&DEAD_WORKERS, 1)
				return
			}
		}
	}
}

func crawlLinksPackage(esClient *elasticsearch.Client, links *[][2]string,
	numberOfWorkers int, numberOfJobs int, lenLinks int) uint64 {

	var wg sync.WaitGroup

	sliceSites := SafeListOfSites{actualSites: make([]Site, 0)}
	sites := make(chan Site, lenLinks+1)
	failedLinks := make(chan map[string]string, lenLinks+1)
	killSignal := make(chan bool)
	finishElasticInsert := make(chan bool)
	done := make(chan bool)

	var proxyIsFree int32
	proxyIsFree = 1

	var numVisitedDomains uint32
	numVisitedDomains = 0

	allParsedLinks := newSafeSetOfLinks()
	packageSize, _ := strconv.Atoi(os.Getenv("NUM_SITES_IN_PACKAGE_SAVE_INDEX"))

	// TODO: if necessary -- set up two goroutince to write in elassticsearch
	var numAddedPages uint64
	go func(finishElasticInsert chan bool, sliceSites *SafeListOfSites, sites <-chan Site,
		allParsedLinks *SafeSetOfLinks, numAddedPages *uint64, done chan bool) {
		wg.Add(1)

		F:
		for true {
			select {
			// take from channel of parsed sites and insert in elasticsearch
			case site := <-sites:
				found := allParsedLinks.checkIfContains(&site.Link)
				if !found {
					allParsedLinks.addLink(&site.Link)
					sliceSites.addSite(&site)
					atomic.AddUint64(numAddedPages, 1)
				}

				if len(sliceSites.actualSites) >= packageSize {
					elasticInsert(esClient, &sliceSites.actualSites, 0)
				}

			default:
				if len(sites) == 0 && int(DEAD_WORKERS) == numberOfWorkers {
					standardLogger.Info("finishElasticInsert break, len(sites) -- ", len(sites))
					break F
				}
			}
		}
		done <- true

		if len(sliceSites.actualSites) != 0 {
			elasticInsert(esClient, &sliceSites.actualSites,0)
		}

		defer wg.Done()
	} (finishElasticInsert, &sliceSites, sites, allParsedLinks, &numAddedPages, done)

	linksQueue := make(chan [2]string, int(NUM_LINKS_PER_CRAWLER * 2))
	for i := 0; i < numberOfWorkers; i++ {
		go crawl(sites, linksQueue, done, killSignal, &wg, failedLinks, i, &proxyIsFree,
			&numVisitedDomains, finishElasticInsert)
	}

	for j := 0; j < numberOfJobs; j++ {
		// TODO: check duplicate at the beginning when take domain

		// avoid http links and complete to a full link of domain
		if !strings.Contains((*links)[j][0], "http") {
			linksQueue <- [2]string{"https://" + (*links)[j][0], (*links)[j][1]}
		} else {
			linksQueue <- (*links)[j]
		}
	}

	//for c := 0; c < numberOfWorkers; c++ {
	<-done
	//}

	standardLogger.Info("collect all elements in done channel")

	close(killSignal)
	close(finishElasticInsert)
	standardLogger.Info("near wg.Wait()")
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

	// load .env file
	//err := godotenv.Load(path.Join("shared_vars.env"))
	err := godotenv.Load(path.Join("..", "shared_vars.env"))

	if err != nil {
		standardLogger.Fatal("Error loading .env file")
	}

	NUM_LINKS_PER_CRAWLER, err = strconv.ParseFloat(os.Getenv("NUM_LINKS_PER_CRAWLER"), 64)
	CUR_INDEX_LINKS_ID = 0
	NUM_INDEXES_LINKS_ID = uint(len(strings.Split(os.Getenv("INDEXES_ELASTIC_LINKS"), " ")))

	// ================== setup configuration ==================
	numberOfWorkers := 8

	// ================== elastic connection ==================
	esClient := elasticConnect()

	insertUkrIdxName := os.Getenv("INDEX_ELASTIC_UKR_COLLECTED_DATA")
	insertRuIdxName := os.Getenv("INDEX_ELASTIC_RU_COLLECTED_DATA")
	titleStr := "start index"
	setIndexFirstId(esClient, insertUkrIdxName, titleStr, "ukr")
	setIndexFirstId(esClient, insertRuIdxName, titleStr, "ru")
	// end elastic connection

	// ================== set last site id in Task Manager ==================
	postBody, _ := json.Marshal(map[string]string{
		"1": "1",
	})
	responseBody := bytes.NewBuffer(postBody)
	http.Post(os.Getenv("TASK_MANAGER_URL")+os.Getenv("TASK_MANAGER_ENDPOINT_SET_LAST_SITE_ID"),
		"application/json", responseBody)

	// ================== get links from TaskManager ==================
	res := responseLinks{}
	getDomainsToParse(&res, false)
	standardLogger.Println("get domains taken: false, parsed: false")

	links := res.Links

	standardLogger.Println("start len(links) -- ", len(links))
	standardLogger.Println("first taken link -- ", links[0])
	standardLogger.Println("last taken link -- ", links[len(links)-1])
	var numberOfJobs = len(links)

	// ------ end getting links from TaskManager

	crawlLinksPackage(esClient, &links, numberOfWorkers, numberOfJobs, len(links))
}
