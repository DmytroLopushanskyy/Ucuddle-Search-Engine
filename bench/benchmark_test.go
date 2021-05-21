package main

import (
	"bufio"
	"encoding/json"
	// "fmt"
	"github.com/gocolly/colly"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
    "testing"
)

type Site struct {
	Title      string    `json:"title"`
	Link       string    `json:"link"`
	Content    string    `json:"inline_text"`
	Hyperlinks []string  `json:hyperlinks`
	AddedAt    time.Time `json:"added_at_time"`
}

type SafeListOfSites struct {
	mu sync.Mutex
	actual_site []Site
}

func (c *SafeListOfSites) addSite(site Site) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.actual_site = append(c.actual_site, site)
	c.mu.Unlock()
}

func (c *SafeListOfSites) checkIfContains(site Site) (int,bool) { 
	c.mu.Lock()
	for i, item := range c.actual_site {
        if item.Link == site.Link {
			c.mu.Unlock()
            return i, true
        }
    }
	c.mu.Unlock()
    return -1, false
}

type SafeList struct {
	mu sync.Mutex
	Links []string
}

func (c *SafeList) addLink(visitedLink string) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.Links = append(c.Links, visitedLink)
	c.mu.Unlock()
}

func (c *SafeList) checkIfContains(link string) (int,bool) { 
	c.mu.Lock()
	for i, item := range c.Links {
        if item == link {
			c.mu.Unlock()
            return i, true
        }
    }
	c.mu.Unlock()
    return -1, false
}

func (c *SafeList) returnLen() int {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	ret := len(c.Links)
	c.mu.Unlock()
	return ret
}

func writeSliceJSON(data []Site, writefile string){
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil{
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

func Find(slice []string, val string) (int, bool) {
    for i, item := range slice {
        if item == val {
            return i, true
        }
    }
    return -1, false
}

func visit_link(lst chan<- Site, link string, visited *SafeList, id int, domain string) (failed error){
	var site Site

	if visited.returnLen() > 10{
		return 
	}

	if !strings.Contains(link, domain){
		return 
	}

	collector := colly.NewCollector(
	)

	collector.OnRequest(func(request *colly.Request) {
		// if (request.URL.String() =="https://oneessencehealing.com/"){
		// fmt.Println("Visiting", request.URL.String())
		// }
	})
	
	collector.OnResponse(func(response *colly.Response) {
		if response.StatusCode != 200 {
			// fmt.Println(response.StatusCode)
		}
	})

	collector.OnError(func(response *colly.Response, err error) {
		failed = err
		if response.StatusCode != 200 && response.StatusCode != 0{
			// fmt.Println(response.StatusCode)
		}
	})

	// var mum map[string][]string
	mum := make(map[string][]string)
	// mum["p"] = make([]string,0)
	// site.Content = ""

	collector.OnHTML("p", func(element *colly.HTMLElement) {
		// fmt.Println(element.Name)
		mum[element.Name] = append(mum[element.Name], element.Text)
		// fmt.Println(mum["p"])
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
		site.Title = site.Title + " " +  element.ChildText("title") + " " + element.DOM.Find("title").Text()
	})

	collector.OnHTML("title",func(element *colly.HTMLElement) {
		site.Title = site.Title + " " + element.Text
	})

	collector.OnHTML("h1",func(element *colly.HTMLElement) {
		site.Title = site.Title + " " + element.Text
	})

	collector.OnHTML("html", func(e *colly.HTMLElement) {
		site.Title = site.Title + " " +  e.ChildAttr(`meta[property="og:title"]`, "content") + " " +  e.ChildText("title") + e.DOM.Find("title").Text()
	})
	
	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if (link != "" ){
			_, found := Find(site.Hyperlinks, link)
			if (!found){
				site.Hyperlinks = append(site.Hyperlinks, link)
			}
		}

	})
	
	// for k := range mum {
    //     fmt.Println(k)
    // }



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

	// fmt.Println(mum["p"])
	

	_, found := visited.checkIfContains(link)

	if !found {
		// *visited = append(*visited, link)
		visited.addLink(link)
		collector.Visit(link)
    }else
	{	
		return
	}


	site.Content = strings.Join(mum["p"], " \n ") + strings.Join(mum["li"], " \n ") + strings.Join(mum["article"], " \n ")

	
	
	if site.Link == ""{
		return 
	}

	// L:
	// for true {
	// 	select {
	// 	case lst <- site:
	// 		break L
	// 	default:
			
	// 	}
	// }
	lst <- site

	
	for _, s := range site.Hyperlinks{
		visit_link(lst, s, visited, id, domain)
	}

	return 
}

func crawl(lst chan<- Site, queue chan string, done, ks chan bool, wg *sync.WaitGroup, failedLinks chan map[string]string, id int) {
	for true {
		select {
		case k := <-queue:
			// site Side 
			
			var failed error 

			// visited := make([]string, 0)
			visited := SafeList{Links: make([]string, 0)}

			failed = visit_link(lst, k, &visited, id, k)
			
			if failed == nil{
				
			}
			
			done <- true
			if failed != nil{
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




func main_crawl(go_rout_num int) {
	links, err := readLines("links.txt")
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	
	var numberOfJobs = len(links)
	var wg sync.WaitGroup
	
	// sliceSites := make([]Site, 0)
	sliceSites := SafeListOfSites{actual_site: make([]Site, 0)}

	
	sites := make(chan Site, len(links)+ 2)


	failedLinks := make(chan map[string]string ,len(links)+ 1)	

	killsignal := make(chan bool)

	// numberOfWritingCrawlers := 3
	// for i := 0; i < numberOfWritingCrawlers; i++ {
	go func(killsignal chan bool, sliceSites *SafeListOfSites, sites <-chan Site){
		F:
		for true{
			select{
			case  l := <-sites:
				_, found := sliceSites.checkIfContains(l)
				if (!found){
					sliceSites.addSite(l)
				}
			case <-killsignal:
				break F
			}
		}
	}(killsignal, &sliceSites, sites)
	// }

	queue := make(chan string)

	done := make(chan bool)

	numberOfWorkers := go_rout_num
	for i := 0; i < numberOfWorkers; i++ {
		go crawl(sites, queue, done, killsignal, &wg, failedLinks,i)
	}

	for j := 0; j < numberOfJobs; j++ {
		
		// select {
		// case k:= queue<-
		// }
		go func(j int) {
			wg.Add(1)
			if !strings.Contains(links[j], "https://"){
				queue <- "https://" + links[j]
			}else{
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




	
	// for {
	// 	val, ok := <-failedLinks
	// 	if ok == false{
	// 		break
	// 	}else{fmt.Println(val["link"])}

	// }
	// for msg := range failedLinks {
	// 	fmt.Println(msg)
	// }

	// writeChannelJSON(sites, "out.json")
	// fmt.Println(len(sliceSites.actual_site))
	writeSliceJSON(sliceSites.actual_site, "out.json")
}


func benchmarkCrawler(i int, b *testing.B) {
    for n := 0; n < b.N; n++ {
		main_crawl(i)
    }
}

func BenchmarkCrawler_20(b *testing.B)  { benchmarkCrawler(20, b) }
func BenchmarkCrawler_40(b *testing.B)  { benchmarkCrawler(40, b) }
func BenchmarkCrawler_60(b *testing.B)  { benchmarkCrawler(60, b) }
func BenchmarkCrawler_80(b *testing.B)  { benchmarkCrawler(80, b) }
func BenchmarkCrawler_100(b *testing.B)  { benchmarkCrawler(100, b) }

func main(){
    main_crawl(2)
}