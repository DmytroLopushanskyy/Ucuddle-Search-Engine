package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/abadojack/whatlanggo"
	"github.com/gocolly/colly"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

func getDomainsToParse(res *responseLinks, ifParseFailedLinks bool) {
	var TM_EndpointLink string
	if !ifParseFailedLinks {
		TM_EndpointLink = os.Getenv("TASK_MANAGER_URL") + os.Getenv("TASK_MANAGER_ENDPOINT_GET_LINKS")
	} else {
		TM_EndpointLink = os.Getenv("TASK_MANAGER_URL") + os.Getenv("TASK_MANAGER_ENDPOINT_GET_FAILED_LINKS")
	}

	var resp *http.Response
	var err error
	waitResponseTime := 0
	for i := 0; i < 5; i++ {
		time.Sleep(time.Duration(waitResponseTime) * time.Second)
		resp, err = http.Get(TM_EndpointLink)

		if err != nil {
			standardLogger.Error("getting response from " + TM_EndpointLink +
				" (iteration ",
				i+1, "): ", err)
		} else {
			break
		}
		waitResponseTime = int(math.Exp(float64(i + 1)))
	}

	standardLogger.Print("Response status get links from Task Manager -- ")
	if resp != nil {
		standardLogger.Println(resp.Status)
	} else {
		standardLogger.Println(resp)
	}

	// check for response error
	if err != nil {
		standardLogger.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		standardLogger.Fatal(err)
	}

	json.Unmarshal(body, &res)
}

func setParsedLink(linkId string) {
	// ------ set link as parsed in TaskManager
	postBody, _ := json.Marshal(map[string]string{
		"parsed_link_id": linkId,
	})
	responseBody := bytes.NewBuffer(postBody)

	var resp *http.Response
	var err error
	waitResponseTime := 0
	for i := 0; i < 5; i++ {
		time.Sleep(time.Duration(waitResponseTime) * time.Second)
		resp, err = http.Post(os.Getenv("TASK_MANAGER_URL") +
			os.Getenv("TASK_MANAGER_ENDPOINT_SET_PARSED_LINK"),
			"application/json",
			responseBody)

		if err != nil {
			standardLogger.Error("Error getting response from TASK_MANAGER_ENDPOINT_SET_PARSED_LINK (iteration ",
				i+1, "): ", err)
		} else {
			break
		}
		waitResponseTime = int(math.Exp(float64(i + 1)))
	}

	// check for response error
	if err != nil {
		standardLogger.Fatal(err)
	}
	defer resp.Body.Close()
}

func findNthSymbol(content *string, symbol string, nOccurrences int) int {
	startPos := 0
	endPos := len(*content)
	pos := -1
	for i := 0; i < nOccurrences; i++ {
		pos = strings.Index((*content)[startPos: endPos], symbol)
		if pos == -1 {
			return -1
		}

		startPos += pos + 1
	}

	return startPos - 1
}

func checkLang(pTagText *string, siteTitle *string) string {
	var content string
	enoughLenChunk := 2000
	lenText := len(*pTagText)
	if lenText == 0 {
		content = *siteTitle
	} else {
		content = *pTagText
	}

	var textChunk string

	// create text chunk to detect language
	if len(content) >= enoughLenChunk {
		slicePos := 0
		nChunks := 5
		chunksLenPosGap := lenText / nChunks
		lenSubChunk := enoughLenChunk / nChunks
		for i := 0; i < nChunks; i++ {
			textChunk = textChunk + content[slicePos:slicePos+lenSubChunk]
			slicePos += chunksLenPosGap
		}
	} else {
		textChunk = content
	}

	textChunk = strings.Join(strings.Fields(textChunk), " ")
	chunkLang := whatlanggo.DetectLang(textChunk)

	return whatlanggo.Langs[chunkLang]
}

func checkMainPageLang(domain string, mainLink *string, visited *SafeSetOfLinks, pageLang *string,
					collector *colly.Collector, mum *map[string][]string,
					site *Site) {
	allUkrLinks := getAllSiteUkrLinks(domain)

	for _, newLink := range allUkrLinks {
		visited.addLink(&newLink)
		collector.Visit(newLink)

		(*site).Content = strings.TrimSpace(strings.Join((*mum)["p"], " \n ") +
			strings.Join((*mum)["li"], " \n ") + strings.Join((*mum)["div"], " \n ") +
			strings.Join((*mum)["article"], " \n "))

		if checkLang(&(*site).Content, &site.Title) == "Ukrainian" {
			*pageLang = "uk"
			*mainLink = newLink
			fmt.Println("checkMainPageLang() Language is Ukrainian", *mainLink)
			break
		}
	}
}

func findCharPos(str *string, char string, numOccurrences int, reversed bool) int {
	lenString := len(*str)

	if !reversed {
		for i := 0; i < lenString; i++ {
			if (*str)[i:i+1] == char {
				numOccurrences--
				if numOccurrences == 0 {
					return i
				}
			}
		}
	} else {
		for i := lenString - 1; i >= 0; i-- {
			if (*str)[i:i+1] == char {
				numOccurrences--
				if numOccurrences == 0 {
					return i
				}
			}
		}
	}

	return -1
}

func getAllSiteUkrLinks(linkToChange string) []string {
	// supported links changes -- www.example.com/uk/, www.example.com/ua/, www.example.de,
	// de.example.com, www.example.com/about/?lang=en,
	// www.example.com.ua, uk-ua.example.com;
	// taken from here -- https://wpml.org/documentation/getting-started-guide/language-setup/language-url-options/

	var arrayLinks []string

	lenLink := len(linkToChange)
	if linkToChange[lenLink-1:lenLink] != "/" {
		fmt.Println(linkToChange[lenLink-1 : lenLink])
		linkToChange += "/"
	}

	arrayLinks = append(arrayLinks, linkToChange)

	if strings.Contains(linkToChange, "https://") {
		// check en. (or other language) in links similar to https://en.wikipedia.org/wiki/
		if linkToChange[10:11] == "." {
			linkToChange = linkToChange[:8] + linkToChange[11:]
		}
	}

	// add new links change
	arrayLinks = append(arrayLinks, linkToChange+"uk/")
	arrayLinks = append(arrayLinks, linkToChange+"ua/")
	arrayLinks = append(arrayLinks, linkToChange+"ukr/")

	// add new link change
	arrayLinks = append(arrayLinks, linkToChange+"?lang=uk/")

	if !strings.Contains(linkToChange, ".ua/") {
		// add new link change
		endSubstringPos := findCharPos(&linkToChange, "/", 3, false)
		sublink := linkToChange[:endSubstringPos]
		startDomainPos := findCharPos(&sublink, ".", 1, true)
		if startDomainPos != -1 {
			arrayLinks = append(arrayLinks, linkToChange[:startDomainPos]+".ua/")
		}

		// add new link change
		wwwPos := strings.Index(linkToChange, "www.")
		if wwwPos != -1 {
			arrayLinks = append(arrayLinks, linkToChange[:wwwPos]+"uk."+linkToChange[wwwPos+4:])
			arrayLinks = append(arrayLinks, linkToChange[:wwwPos]+"ua."+linkToChange[wwwPos+4:])
		} else {
			arrayLinks = append(arrayLinks, linkToChange[:8]+"ua."+linkToChange[8:])
			arrayLinks = append(arrayLinks, linkToChange[:8]+"uk."+linkToChange[8:])
		}

		// add new link change
		endDomainPos := strings.Index(linkToChange, ".com/")
		if endDomainPos != -1 {
			arrayLinks = append(arrayLinks, linkToChange[:endDomainPos+4]+".ua/"+linkToChange[endDomainPos+5:])
		}

		// add new link change
		if wwwPos != -1 {
			arrayLinks = append(arrayLinks, linkToChange[:wwwPos]+"uk-ua."+linkToChange[wwwPos+4:])
		}
	}

	return arrayLinks
}