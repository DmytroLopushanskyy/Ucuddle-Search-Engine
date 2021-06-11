package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Site struct {
	SiteId     uint64    `json:"site_id"`
	Title      string    `json:"title"`
	Link       string    `json:"link"`
	PageRank   uint64    `json:"page_rank"`
	Content    string    `json:"content"`
	Hyperlinks []string  `json:"hyperlinks"`
	AddedAt    time.Time `json:"added_at_time"`
}

type responseLastSiteId struct {
	LastSiteId uint64 `json:"last_site_id"`
}

func updateIndexMapping(es *elasticsearch.Client, indexName string) {
	{
		res, err := es.Indices.Close([]string{indexName})
		fmt.Println(res, err)
		if err != nil { // SKIP
			log.Printf("Error getting the response: %s", err) // SKIP
		} // SKIP
		defer res.Body.Close() // SKIP
	}

	{
		var buf bytes.Buffer
		query := map[string]interface{}{
			"properties": map[string]interface{}{
				"site_id": map[string]interface{}{
					"type": "unsigned_long",
				},
				"!!!my_site_id": map[string]interface{}{ // TODO delete
					"type": "unsigned_long",
				},
				"content": map[string]interface{}{
					"type":                  "text",
					"analyzer":              "english",
					"search_analyzer":       "english",
					"search_quote_analyzer": "english",
				},
				"link": map[string]interface{}{
					"type": "text",
				},
				"added_at_time": map[string]interface{}{
					"type": "date_nanos",
				},
			},
		}

		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			log.Printf("Error encoding query: %s", err)
		}

		res, err := es.Indices.PutMapping([]string{indexName}, &buf)
		fmt.Println(res, err)
		if err != nil { // SKIP
			log.Printf("Error getting the response: %s", err) // SKIP
		} // SKIP
		defer res.Body.Close() // SKIP
	}

	{
		res, err := es.Indices.Open([]string{indexName})
		fmt.Println(res, err)
		if err != nil { // SKIP
			log.Printf("Error getting the response: %s", err) // SKIP
		} // SKIP
		defer res.Body.Close() // SKIP
	}

	fmt.Println("Settings updated !!!")
}

func getIndexMapping(es *elasticsearch.Client, indexName string) {
	res, err := es.Indices.GetMapping(es.Indices.GetMapping.WithIndex(indexName))
	fmt.Println(res, err)
	if err != nil { // SKIP
		log.Printf("Error getting the response: %s", err) // SKIP
	} // SKIP
	defer res.Body.Close() // SKIP
}

func createIndexForLinks(es *elasticsearch.Client, saveStrIdx string) {
	var buf bytes.Buffer

	query := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"link_id": map[string]interface{}{
					"type": "unsigned_long",
				},
				"link": map[string]interface{}{
					"type": "text",
				},
				"taken": map[string]interface{}{
					"type": "boolean",
				},
				"parsed": map[string]interface{}{
					"type": "boolean",
				},
				"added_at_time": map[string]interface{}{
					"type": "date_nanos",
				},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Printf("Error encoding query: %s", err)
	}

	var res *esapi.Response
	res, err := es.Indices.Create(
		saveStrIdx,
		es.Indices.Create.WithBody(&buf),
	)

	fmt.Println("\nsetIndexUkrAnalyzer")
	fmt.Println(res)

	if res.Status() != "200 OK" { // SKIP
		fmt.Println("ERROR in setIndexUkrAnalyzer():")
		fmt.Println(res, err)
		os.Exit(3)
	}

	defer res.Body.Close()
}

func setIndexUkrAnalyzer(es *elasticsearch.Client, saveStrIdx string) {
	var buf bytes.Buffer

	//lang := "ukrainian"
	lang := "english"
	query := map[string]interface{}{
		"settings": map[string]interface{}{
			"index": map[string]interface{}{
				"number_of_shards": 2,
			},
			"analysis": map[string]interface{}{
				"analyzer": lang,
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"site_id": map[string]interface{}{
					"type": "unsigned_long",
				},
				"title": map[string]interface{}{
					"type":                  "text",
					"analyzer":              lang,
					"search_analyzer":       lang,
					"search_quote_analyzer": lang,
				},
				"content": map[string]interface{}{
					"type":                  "text",
					"analyzer":              lang,
					"search_analyzer":       lang,
					"search_quote_analyzer": lang,
				},
				"link": map[string]interface{}{
					"type": "text",
				},
				"added_at_time": map[string]interface{}{
					"type": "date_nanos",
				},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		fmt.Printf("Error encoding query: %s", err)
	}

	var res *esapi.Response
	res, err := es.Indices.Create(
		saveStrIdx,
		es.Indices.Create.WithBody(&buf),
	)

	fmt.Println("\nsetIndexUkrAnalyzer")
	fmt.Println(res)

	if res.Status() != "200 OK" { // SKIP
		fmt.Println("ERROR in setIndexUkrAnalyzer():")
		fmt.Println(res, err)
		os.Exit(3)
	}

	defer res.Body.Close()
}

func setIndexFirstId(es *elasticsearch.Client, idxName string,
	titleStr string) {
	var dataArr []Site

	res, err := es.Indices.Get([]string{idxName})
	if err != nil { // SKIP
		fmt.Printf("Error getting the response: %s", err)
	} // SKIP
	defer res.Body.Close() // SKIP

	if res.Status() != "200 OK" {
		fmt.Println("\n\n ========== Index does not exist")

		site := Site{
			Title: titleStr[:len(titleStr)],
			Link:  "https:",
		}

		dataArr = append(dataArr, site)

		if os.Getenv("DEBUG") == "true" {
			fmt.Printf("%v", dataArr)
		}

		setIndexUkrAnalyzer(es, idxName)
		elasticInsert(es, &dataArr, &idxName, 1)
	} else {
		fmt.Println("\n\n ========== Index already exists")
	}
}

func elasticInsert(es *elasticsearch.Client, dataArr *[]Site, saveStrIdx *string,
	externalLastId uint64) {

	var (
		wg sync.WaitGroup
	)

	var mu sync.Mutex
	var curIndexLastId uint64

	if externalLastId == 0 {
		// ------ get last_site_id from TaskManager

		var resp *http.Response
		var err error
		waitResponseTime := 0
		for i := 0; i < 5; i++ {
			time.Sleep(time.Duration(waitResponseTime) * time.Second)
			resp, err = http.Get(os.Getenv("TASK_MANAGER_URL") + os.Getenv("TASK_MANAGER_ENDPOINT_GET_LAST_SITE_ID"))

			if err != nil {
				fmt.Println("getting response to TASK_MANAGER_ENDPOINT_GET_LAST_SITE_ID (iteration ",
					i+1, "): ", err)
			} else {
				break
			}
			waitResponseTime = int(math.Exp(float64(i + 1)))
		}

		// check for response error
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		res := responseLastSiteId{}
		json.Unmarshal(body, &res)

		curIndexLastId = res.LastSiteId

	} else {
		curIndexLastId = externalLastId
	}
	fmt.Println("curIndexLastId -- ", curIndexLastId)

	// Index documents concurrently
	//
	for i, site := range *dataArr {
		wg.Add(1)

		go func(nDoc int, site2 Site, mu *sync.Mutex, indexLastId *uint64) {
			defer wg.Done()

			mu.Lock()
			site2.SiteId = *indexLastId
			*indexLastId++
			mu.Unlock()
			//atomic.AddUint64(indexLastId, 1)

			site2.AddedAt = time.Now().UTC()

			// Build the request body.
			res2B, _ := json.Marshal(site2)

			// Set up the request object.
			req := esapi.IndexRequest{
				Index:      *saveStrIdx,
				DocumentID: strconv.FormatUint(site2.SiteId, 10),
				Body:       strings.NewReader(string(res2B)),
				Refresh:    "true",
			}

			// Perform the request with the client.
			var res *esapi.Response
			var err error

			waitResponseTime := 0
			for j := 0; j < 5; j++ {
				time.Sleep(time.Duration(waitResponseTime) * time.Second)

				// for testing
				//if i == 2 {
				res, err = req.Do(context.Background(), es)
				//} else {
				//	err = errors.New("test error")
				//}

				if err != nil {
					fmt.Errorf("getting response (iteration ", j+1, "): ", err)
				} else {
					break
				}
				waitResponseTime = int(math.Exp(float64(j + 1)))
			}

			if err != nil {
				fmt.Printf("Error getting response: %s", err)
			}
			defer res.Body.Close()

			if res.IsError() {
				fmt.Errorf("[%s] Error indexing document ID=%d", res.Status(), i+1)
				fmt.Println("response -- ", res)
			}
		}(i, site, &mu, &curIndexLastId)

		fmt.Println(site.Link, " inserted")
	}
	wg.Wait()

	mu.Lock()
	*dataArr = (*dataArr)[:0]
	mu.Unlock()

	fmt.Println(strings.Repeat("-", 37))
}