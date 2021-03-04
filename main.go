package main

import ("fmt"
		"encoding/json"
		"time"
		"log"
		"github.com/gocolly/colly"
		"io/ioutil"
		)

type Fact struct {
	Description string `json:"desctiption"`
}

func writeJSON(data []Fact, writefile string){
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil{
		log.Println("error")
		return
	}

	_ = ioutil.WriteFile(writefile, file, 0644)
}

func crawl(link string, writefile string){
	allFacts := make([]Fact, 0)

    // colector -> colly scrape interface
	collector := colly.NewCollector()


	// happens on the request
	collector.OnRequest(func (request *colly.Request){
		fmt.Println("Visiting", request.URL.String())
	})

	// Called right after OnResponse
	collector.OnHTML("p", func(element *colly.HTMLElement){
		fact := Fact{ Description: element.Text,}

		allFacts = append(allFacts, fact)

	})
	

	collector.Visit(link)

	writeJSON(allFacts, writefile)
}

func main() {
	var links [3]string;
	links[0] = "https://en.wikipedia.org/wiki/Werner_Heisenberg"
	links[1] = "https://en.wikipedia.org/wiki/Warren_G._Harding"
	links[2] = "https://en.wikipedia.org/wiki/World_of_Warcraft"



	p := fmt.Println

	then_1 := time.Now()
	
	crawl(links[0],"out_1.json")
	crawl(links[1],"out_2.json")
	crawl(links[2],"out_3.json")

	then_2 := time.Now()
	p("")
	p("consecutive crawl time:")
	p(then_2.Sub(then_1))
	p("")
	

	then_3 := time.Now()
	

	// does not work properly yet
	go crawl(links[0], "out_1.json")
	go crawl(links[1], "out_2.json")
	go crawl(links[2], "out_3.json")

	then_4 := time.Now()
	p("parallel crawl time:")
	p(then_4.Sub(then_3))
	p("")


}