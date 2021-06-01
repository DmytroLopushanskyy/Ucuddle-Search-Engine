package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"os"
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