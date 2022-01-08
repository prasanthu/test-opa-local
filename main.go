package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

var logger = log.Default()

func getInputData(filename string) []byte {
	inputFile, err := os.Open(filename)
	if err != nil {
		logger.Panic(err)
	}
	defer inputFile.Close()
	data, err := ioutil.ReadAll(inputFile)
	if err != nil {
		logger.Panic(err)
	}
	return data
}

func callLocalOpa(client *http.Client, inputData []byte) map[string]interface{} {
	opaURL := "http://localhost:8181/v1/data/envoy"

	req, err := http.NewRequest("POST", opaURL, strings.NewReader(string(inputData)))
	if err != nil {
		logger.Panic("Unable to create http request", err)
	}
	traceID := uuid.NewString()
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Trace-Id", traceID)

	resp, err := client.Do(req)
	if err != nil {
		logger.Panic("Unable to perform http request", err)
	}
	defer resp.Body.Close()

	//logger.Printf("Status code %d Trace Id %s \n", resp.StatusCode, traceID)

	if resp.StatusCode != 200 {
		bodyVal, _ := ioutil.ReadAll(resp.Body)
		logger.Println(string(bodyVal))
		logger.Panic(fmt.Sprintf("Expected status code 200, not: %v body: \n%v", resp.StatusCode, string(bodyVal)))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Panic("Unable to parse body")
	}

	var opaReply map[string]interface{}
	err = json.Unmarshal(body, &opaReply)
	if err != nil {
		logger.Panic("Unable to unmarshal body")
	}
	return opaReply
}

func doAPITests(client *http.Client, inputData []byte, count int) {
	for j := 0; j < count; j++ {
		t := time.Now()
		callLocalOpa(client, inputData)
		d := time.Since(t)
		if d/time.Millisecond > 3 {
			logger.Printf("Time > 3 ms %v\n", d)
		}
	}
}

func main() {
	client := &http.Client{}
	inputData := getInputData("input.json")
	doAPITests(client, inputData, 5000)
}
