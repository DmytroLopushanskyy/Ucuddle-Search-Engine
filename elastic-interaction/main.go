package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/joho/godotenv"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

func elasticConnect() *elasticsearch.Client {
	fmt.Println("start connecting")
	log.SetFlags(0)

	var (
		r map[string]interface{}
	)

	// Initialize a client with the default mapping.
	//
	// An `ELASTICSEARCH_URL` environment variable will be used when exported.
	//
	cfg := elasticsearch.Config{
		Username: os.Getenv("Username"),
		Password: os.Getenv("Password"),
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// Get cluster info
	//
	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print client and server version numbers.
	log.Printf("Client: %s", elasticsearch.Version)
	log.Printf("Server: %s", r["version"].(map[string]interface{})["number"])
	log.Println(strings.Repeat("~", 37))

	return es
}

func searchQuery(es *elasticsearch.Client, searchStrIdx string, queryBuf *bytes.Buffer) []gjson.Result {
	// Search for the indexed documents with full index name and word in titles
	//

	// Perform the search request.
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(searchStrIdx),
		es.Search.WithBody(queryBuf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)

	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
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
	log.Printf(
		"[%s] %d hits; took: %dms\n",
		res.Status(),
		values[0].Int(),
		values[1].Int(),
	)

	return values[2].Array()
}

func indexGetLastId(esClient *elasticsearch.Client, indexName string,
	nLastRecords int) uint64 {
	// Build the request body.
	var buf bytes.Buffer

	// query for retrieving the last id in indexName by "added_at_time" parameter
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},

		"size": nLastRecords,

		// TODO
		"sort": map[string]interface{}{
			"site_id": map[string]interface{}{
				"order": "desc",
			},
		},

		//"track_total_hits": false,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	results := searchQuery(esClient, indexName, &buf)
	//fmt.Println("result", results)
	for i, result := range results {
		fmt.Println(i, result.Map()["_source"].Map()["title"])
		fmt.Println("_id", result.Map()["_id"])
		fmt.Println("site_id -- ", result.Map()["_source"].Map()["site_id"])
		fmt.Println(result.Map()["_source"].Map()["added_at_time"])
		fmt.Println("\n")
	}

	lastIdx := results[0].Map()
	return lastIdx["_source"].Map()["site_id"].Uint()
}

func main() {
	// Perform health-check
	//for {
	//	resp, err := http.Get("http://elasticsearch:9200")
	//	if err == nil {
	//		break
	//	}
	//	time.Sleep(time.Second)
	//}
	// Elasticsearch server has started. The program begins

	// TODO: how
	// to read line from user without \n with normal construction

	fmt.Println("Application started")

	err := godotenv.Load(path.Join("..", "crawlers-env.env"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Before this module execution, run elasticsearch and kibana servers on your computer
	esClient := elasticConnect()

	for {
		fmt.Println("\n\nEnter function number to execute: ")
		functionNames := []string{"insert words", "search indexes", "delete all docs in index",
			"get last id and print n last document titles in the index",
			"update mapping in index", "get index mapping",
			"delete several indexes"}
		for i, funcName := range functionNames {
			fmt.Println(i+1, " -- ", funcName)
		}

		var input string
		//var err error

		fmt.Scanln(&input, &err)
		if err != nil {
			log.Fatalf("Error getting response: %s", err)
		}

		if input == "1" {
			fmt.Println("Enter your index name for saving in database")
			reader := bufio.NewReader(os.Stdin)
			idxName, _ := reader.ReadString('\n')
			idxName = idxName[:len(idxName)-1]

			fmt.Println("Enter your data for indexing")

			var dataArr []string
			continueInput := 1
			for continueInput == 1 {
				fmt.Println("Enter your textLine: ")
				reader := bufio.NewReader(os.Stdin)
				textLine, _ := reader.ReadString('\n')

				if textLine == "q\n" {
					continueInput = 0
					break
				}

				dataArr = append(dataArr, textLine[:len(textLine)-1])
			}

			fmt.Printf("%v", dataArr)
			indexing(esClient, dataArr, idxName)

		} else if input == "2" {
			fmt.Println("Enter your index name for searching")
			reader := bufio.NewReader(os.Stdin)
			idxName, _ := reader.ReadString('\n')
			idxName = idxName[:len(idxName)-1]

			fmt.Println("Enter your title name for searching")
			reader = bufio.NewReader(os.Stdin)
			titleName, _ := reader.ReadString('\n')
			titleName = titleName[:len(titleName)-1]

			searching(esClient, idxName, titleName)

		} else if input == "3" {
			fmt.Println("Enter your index for deleting")
			reader := bufio.NewReader(os.Stdin)
			idxName, _ := reader.ReadString('\n')
			idxName = idxName[:len(idxName)-1]

			fmt.Println("Enter your document id in this index for deleting",
				"\n or '--' to delete all documents in this index")
			reader = bufio.NewReader(os.Stdin)
			idStr, _ := reader.ReadString('\n')
			idStr = idStr[:len(idStr)-1]

			deleting(esClient, idxName, idStr)

		} else if input == "4" {
			fmt.Println("Enter your index to get last document id and print n last document titles")
			reader := bufio.NewReader(os.Stdin)
			idxName, _ := reader.ReadString('\n')
			idxName = idxName[:len(idxName)-1]
			//insertIdxName := "t_english_sites-a16"

			fmt.Println("Enter number of the last elements to get")
			reader = bufio.NewReader(os.Stdin)
			nLastRecordsStr, _ := reader.ReadString('\n')
			nLastRecords, _ := strconv.Atoi(nLastRecordsStr[:len(nLastRecordsStr)-1])

			indexLastIdInt := indexGetLastId(esClient, idxName, nLastRecords)
			fmt.Println("indexLastIdInt -- ", indexLastIdInt)

		} else if input == "5" {
			fmt.Println("Enter your index to update mapping")
			reader := bufio.NewReader(os.Stdin)
			idxName, _ := reader.ReadString('\n')
			idxName = idxName[:len(idxName)-1]

			updateIndexMapping(esClient, idxName)

		} else if input == "6" {
			fmt.Println("Enter your index to get mapping")
			reader := bufio.NewReader(os.Stdin)
			idxName, _ := reader.ReadString('\n')
			idxName = idxName[:len(idxName)-1]

			getIndexMapping(esClient, idxName)

		} else if input == "7" {
			fmt.Println("Enter indexes to delete (input indexes separated by one space)")
			reader := bufio.NewReader(os.Stdin)
			indexes, _ := reader.ReadString('\n')
			indexes = indexes[:len(indexes)-1]

			deleteIndexes(esClient, strings.Split(indexes, " "))

		}
	}
}
