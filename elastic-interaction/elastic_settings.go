package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"log"
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