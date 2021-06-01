package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"log"
	"os"
)

func updateIndexMapping(es *elasticsearch.Client, indexName string) {
	{
		res, err := es.Indices.Close([]string{indexName})
		fmt.Println(res, err)
		if err != nil { // SKIP
			log.Fatalf("Error getting the response: %s", err) // SKIP
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
			log.Fatalf("Error encoding query: %s", err)
		}

		res, err := es.Indices.PutMapping([]string{indexName}, &buf)
		fmt.Println(res, err)
		if err != nil { // SKIP
			log.Fatalf("Error getting the response: %s", err) // SKIP
		} // SKIP
		defer res.Body.Close() // SKIP
	}

	{
		res, err := es.Indices.Open([]string{indexName})
		fmt.Println(res, err)
		if err != nil { // SKIP
			log.Fatalf("Error getting the response: %s", err) // SKIP
		} // SKIP
		defer res.Body.Close() // SKIP
	}

	fmt.Println("Settings updated !!!")
}

func getIndexMapping(es *elasticsearch.Client, indexName string) {
	res, err := es.Indices.GetMapping(es.Indices.GetMapping.WithIndex(indexName))
	fmt.Println(res, err)
	if err != nil { // SKIP
		log.Fatalf("Error getting the response: %s", err) // SKIP
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
		log.Fatalf("Error encoding query: %s", err)
	}

	var res *esapi.Response
	res, err := es.Indices.Create(
		saveStrIdx,
		es.Indices.Create.WithBody(&buf),
	)

	fmt.Println("\nsetIndexAnalyzer")
	fmt.Println(res)

	if res.Status() != "200 OK" { // SKIP
		fmt.Println("ERROR in setIndexAnalyzer():")
		fmt.Println(res, err)
		os.Exit(3)
	}

	defer res.Body.Close()
}
