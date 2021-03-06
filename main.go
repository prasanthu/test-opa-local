package main

import (
	"bytes"
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

func marshalToBytes(v interface{}) []byte {
	payloadBytes, err := json.Marshal(v)
	if err != nil {
		logger.Panic("Failed to json marshal payload", err)
	}
	return payloadBytes
}

func getPretty(body []byte) string {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		logger.Panic("Unable to prettify the json payload", string(body), err)
	}
	return prettyJSON.String()
}

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
	opaURL := "http://localhost:8181/v1/data/envoy?metrics=true"

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
		result := callLocalOpa(client, inputData)
		d := time.Since(t)
		if d/time.Millisecond > 5 {
			logger.Printf("Time > 5 ms %v\n", d)
			logger.Printf("OPA Reply  %+v\n", getPretty(marshalToBytes(result["metrics"])))
		}
	}
}

func main() {
	client := &http.Client{}
	inputData := getInputData("input.json")
	doAPITests(client, inputData, 5000)
}
