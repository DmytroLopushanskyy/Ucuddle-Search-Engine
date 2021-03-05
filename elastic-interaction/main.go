package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func connect() *elasticsearch.Client {
	fmt.Println("start connecting")
	log.SetFlags(0)

	var (
		r map[string]interface{}
	)

	// Initialize a client with the default settings.
	//
	// An `ELASTICSEARCH_URL` environment variable will be used when exported.
	//
	es, err := elasticsearch.NewDefaultClient()
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

func indexing(es *elasticsearch.Client, dataArr []string, saveStrIdx string) {
	var (
		wg sync.WaitGroup
	)

	// Index documents concurrently
	//
	for i, title := range dataArr {
		wg.Add(1)

		go func(i int, title string) {
			defer wg.Done()

			// Build the request body.
			var b strings.Builder
			b.WriteString(`{"title" : "`)
			b.WriteString(title)
			b.WriteString(`"}`)

			// Set up the request object.
			req := esapi.IndexRequest{
				Index:      saveStrIdx,
				DocumentID: strconv.Itoa(i + 1),
				Body:       strings.NewReader(b.String()),
				Refresh:    "true",
			}

			// Perform the request with the client.
			res, err := req.Do(context.Background(), es)
			fmt.Println(res, err)

			if err != nil {
				log.Fatalf("Error getting response: %s", err)
			}
			defer res.Body.Close()

			if res.IsError() {
				log.Printf("[%s] Error indexing document ID=%d", res.Status(), i+1)
			} else {
				// Deserialize the response into a map.
				var r map[string]interface{}
				if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
					log.Printf("Error parsing the response body: %s", err)
				} else {
					// Print the response status and indexed document version.
					log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
				}
			}
		}(i, title)
	}
	wg.Wait()

	log.Println(strings.Repeat("-", 37))
}

func searching(es *elasticsearch.Client, searchStrIdx string, titleName string) {
	// Search for the indexed documents with full index name and word in titles
	//
	// Build the request body.
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"title": titleName,
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	fmt.Println(searchStrIdx)
	fmt.Println(len(searchStrIdx))

	// Perform the search request.
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(searchStrIdx),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	fmt.Println(res, err)

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

	var (
		r map[string]interface{}
	)

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}

	log.Println(strings.Repeat("=", 37))
}

func deleting(es *elasticsearch.Client, deleteStrIdx string, id string) {
	// Delete documents by id and index name

	// Perform the delete request.
	var res *esapi.Response
	var err error
	if id != "--" {
		res, err = es.Delete(deleteStrIdx, id)
	} else {
		fmt.Println("id == --")

		res, err = es.DeleteByQuery(
			[]string{deleteStrIdx},
			strings.NewReader(`{
				  "query": {
					"match_all": {}
				  }
				}`),
			es.DeleteByQuery.WithConflicts("proceed"),
		)

		fmt.Println(res, err)
	}

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
}

func main() {
	time.Sleep(10 * time.Second)
	// TODO: how
	// to read line from user without \n with normal construction

	fmt.Println("Application started")

	// Before this module execution, run elasticsearch and kibana servers on your computer
	es := connect()

	fmt.Println("Enter function number to execute: ")
	functionNames := []string{"insert words", "search indexes", "delete indexes"}
	for i, funcName := range functionNames {
		fmt.Println(i+1, " -- ", funcName)
	}

	var input string
	var err error

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
		indexing(es, dataArr, idxName)

	} else if input == "2" {
		fmt.Println("Enter your index name for searching")
		reader := bufio.NewReader(os.Stdin)
		idxName, _ := reader.ReadString('\n')
		idxName = idxName[:len(idxName)-1]

		fmt.Println("Enter your title name for searching")
		reader = bufio.NewReader(os.Stdin)
		titleName, _ := reader.ReadString('\n')
		titleName = titleName[:len(titleName)-1]

		searching(es, idxName, titleName)

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

		deleting(es, idxName, idStr)
	}
}
