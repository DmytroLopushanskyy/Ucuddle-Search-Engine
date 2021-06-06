package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)


func elasticConnect() *elasticsearch.Client {
	fmt.Println("start connecting")

	var (
		r map[string]interface{}
	)

	// Initialize a client with the default settings.
	//
	// An `ELASTICSEARCH_URL` environment variable will be used when exported.
	//
	cfg := elasticsearch.Config{
		Username: os.Getenv("Username"),
		Password: os.Getenv("Password"),
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		standardLogger.Fatalf("Error creating the client: %s", err)
	}

	// Get cluster info
	//
	res, err := es.Info()
	if err != nil {
		standardLogger.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		standardLogger.Fatalf("Error: %s", res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		standardLogger.Fatalf("Error parsing the response body: %s", err)
	}
	// Print client and server version numbers.
	standardLogger.Printf("Client: %s", elasticsearch.Version)
	standardLogger.Printf("Server: %s", r["version"].(map[string]interface{})["number"])
	fmt.Println(strings.Repeat("~", 37))

	return es
}

func setIndexUkrAnalyzer(es *elasticsearch.Client, saveStrIdx string) {
	var buf bytes.Buffer

	lang := "ukrainian"
	query := map[string]interface{}{
		"settings": map[string]interface{}{
			"analysis": map[string]interface{}{
				"analyzer": lang,
			},
			"index": map[string]interface{}{
				"number_of_shards": 3,
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
		standardLogger.Fatalf("Error encoding query: %s", err)
	}

	var res *esapi.Response
	res, err := es.Indices.Create(
		saveStrIdx,
		es.Indices.Create.WithBody(&buf),
	)

	standardLogger.Println("\nsetIndexUkrAnalyzer")
	fmt.Println(res)

	if res.Status() != "200 OK" { // SKIP
		fmt.Println("ERROR in setIndexUkrAnalyzer():")
		fmt.Println(res, err)
		os.Exit(3)
	}

	defer res.Body.Close()
}

func setIndexRuAnalyzer(es *elasticsearch.Client, saveStrIdx string) {
	var buf bytes.Buffer

	lang := "russian"
	query := map[string]interface{}{
		"settings": map[string]interface{}{
			"index": map[string]interface{}{
				"number_of_shards": 3,
			},
			"analysis": map[string]interface{}{
				"filter": map[string]interface{}{
					"russian_stop": map[string]interface{}{
						"type":       "stop",
						"stopwords":  "_russian_",
					},
					"russian_stemmer": map[string]interface{}{
						"type":       "stemmer",
						"language":   lang,
					},
				},
				"analyzer": map[string]interface{}{
					"rebuilt_russian": map[string]interface{}{
						"tokenizer":  "standard",
						"filter": []string{
							"lowercase",
							"russian_stop",
							"russian_stemmer",
						},
					},
				},
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
		standardLogger.Fatalf("Error encoding query: %s", err)
	}

	var res *esapi.Response
	res, err := es.Indices.Create(
		saveStrIdx,
		es.Indices.Create.WithBody(&buf),
	)

	standardLogger.Println("\nsetIndexUkrAnalyzer")
	fmt.Println(res)

	if res.Status() != "200 OK" { // SKIP
		fmt.Println("ERROR in setIndexUkrAnalyzer():")
		fmt.Println(res, err)
		os.Exit(3)
	}

	defer res.Body.Close()
}

func setIndexFirstId(es *elasticsearch.Client, idxName string,
	titleStr string, lang string) {
	var dataArr []Site

	res, err := es.Indices.Get([]string{idxName})
	if err != nil { // SKIP
		standardLogger.Fatalf("Error getting the response: %s", err)
	} // SKIP
	defer res.Body.Close() // SKIP

	if res.Status() != "200 OK" {
		fmt.Println("\n\n ========== Index does not exist")

		site := Site{
			Title: titleStr[:len(titleStr)],
			Link:  "https:",
		}

		if os.Getenv("DEBUG") == "true" {
			fmt.Printf("%v", dataArr)
		}

		if lang == "ukr" {
			site.Lang = "ukr"
			setIndexUkrAnalyzer(es, idxName)
		} else if lang == "ru" {
			site.Lang = "ru"
			setIndexRuAnalyzer(es, idxName)
		}
		dataArr = append(dataArr, site)

		elasticInsert(es, &dataArr, 1)
	} else {
		fmt.Println("\n\n ========== Index already exists")
	}
}

func elasticInsert(es *elasticsearch.Client, dataArr *[]Site, externalLastId uint64) {
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
			resp, err = http.Get(os.Getenv("TASK_MANAGER_URL") +
				os.Getenv("TASK_MANAGER_ENDPOINT_GET_LAST_SITE_ID"))

			if err != nil {
				standardLogger.Error("getting response from TASK_MANAGER_ENDPOINT_GET_LAST_SITE_ID (iteration ",
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

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			standardLogger.Fatal(err)
		}

		res := responseLastSiteId{}
		json.Unmarshal(body, &res)

		curIndexLastId = res.LastSiteId

	} else {
		curIndexLastId = externalLastId
	}
	standardLogger.Debug("curIndexLastId -- ", curIndexLastId)

	mapInsertIdxName := make(map[string]string)
	mapInsertIdxName["ukr"] = os.Getenv("INDEX_ELASTIC_UKR_COLLECTED_DATA")
	mapInsertIdxName["ru"] = os.Getenv("INDEX_ELASTIC_RU_COLLECTED_DATA")

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

			site2.AddedAt = time.Now().UTC()

			// Build the request body.
			res2B, _ := json.Marshal(site2)

			// Set up the request object.
			req := esapi.IndexRequest{
				Index:      mapInsertIdxName[site2.Lang],
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

				res, err = req.Do(context.Background(), es)

				if err != nil {
					standardLogger.Error("getting response (iteration ", j+1, "): ", err)
				} else {
					break
				}
				waitResponseTime = int(math.Exp(float64(j + 1)))
			}

			if err != nil {
				standardLogger.Fatalf("Error getting response: %s", err)
			}
			defer res.Body.Close()

			if res.IsError() {
				standardLogger.Errorf("[%s] Error indexing document ID=%d", res.Status(), i+1)
				standardLogger.Println("response -- ", res)
			}
		}(i, site, &mu, &curIndexLastId)

		standardLogger.Println(site.Link, " inserted")
	}
	wg.Wait()

	mu.Lock()
	*dataArr = (*dataArr)[:0]
	mu.Unlock()

	fmt.Println(strings.Repeat("-", 37))
}

func indexGetLastId(esClient *elasticsearch.Client, indexName string) uint64 {
	// Build the request body.
	var buf bytes.Buffer

	// query for retrieving the last id in indexName by "added_at_time" parameter
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},

		"size": 1,

		"sort": map[string]interface{}{
			"site_id": map[string]interface{}{
				"order": "desc",
			},
		},

		//"track_total_hits": false,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		standardLogger.Fatalf("Error encoding query: %s", err)
	}

	results := searchQuery(esClient, indexName, &buf)
	for i, result := range results {
		fmt.Println(i, result.Map()["_source"].Map()["title"])
		fmt.Println("_id", result.Map()["_id"])
		fmt.Println("site_id -- ", result.Map()["_source"].Map()["site_id"])
		fmt.Println(result.Map()["_source"].Map()["added_at_time"])
	}

	lastIdx := results[0].Map()
	return lastIdx["_source"].Map()["site_id"].Uint()
}

func searchQuery(es *elasticsearch.Client, searchStrIdx string, queryBuf *bytes.Buffer) []gjson.Result {
	// Search for the indexed documents with full index name and word in titles
	//

	// Perform the search request.
	var res *esapi.Response
	var err error

	waitResponseTime := 0

	for i := 0; i < 5; i++ {
		time.Sleep(time.Duration(waitResponseTime) * time.Second)
		res, err = es.Search(
			es.Search.WithContext(context.Background()),
			es.Search.WithIndex(searchStrIdx),
			es.Search.WithBody(queryBuf),
			es.Search.WithTrackTotalHits(true),
		)

		if err != nil {
			standardLogger.Error("getting response (iteration ", i+1, "): ", err)
		} else {
			break
		}
		waitResponseTime = int(math.Exp(float64(i + 1)))
	}

	if err != nil {
		standardLogger.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			standardLogger.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			standardLogger.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	var b bytes.Buffer
	b.ReadFrom(res.Body)

	// usage of gjson lib for easily parsing res.Body json
	values := gjson.GetManyBytes(b.Bytes(), "hits.total.value", "took", "hits.hits")

	// Print the response status, number of results, and request duration.
	standardLogger.Printf(
		"[%s] %d hits; took: %dms\n",
		res.Status(),
		values[0].Int(),
		values[1].Int(),
	)

	return values[2].Array()
}
