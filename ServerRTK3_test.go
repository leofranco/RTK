package main

import (
    "net/http"
    "testing"
    "log"
    "io/ioutil"
)

func TestSet(t *testing.T) {
    var setEndPoint string = "http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set"

    response, err := http.Get(setEndPoint)
    if err != nil && response == nil {
       log.Fatalf("Error sending request to API endpoint. %+v", err)
    } else {
        // Close the connection to reuse it
        defer response.Body.Close()

	if response.StatusCode != 200 {
	   log.Fatalf("Request was sent but status code is not 200, StatusCode ", response.StatusCode)
	} else {
	    body, err := ioutil.ReadAll(response.Body)
            if err != nil {
                log.Fatalf("Couldn't parse response body. %+v", err)
            }
	    log.Println("Response Body:", string(body))    
        }
    }
}

func TestGet(t *testing.T) {
    var setEndPoint string = "http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/get"

    response, err := http.Get(setEndPoint)
    if err != nil && response == nil {
       log.Fatalf("Error sending request to API endpoint. %+v", err)
    } else {
        // Close the connection to reuse it
        defer response.Body.Close()

        if response.StatusCode != 200 {
           log.Fatalf("Request was sent but status code is not 200, StatusCode ", response.StatusCode)
        } else {
            body, err := ioutil.ReadAll(response.Body)
            if err != nil {
                log.Fatalf("Couldn't parse response body. %+v", err)
            }
            log.Println("Response Body:", string(body))
        }
    }
}

func TestNot(t *testing.T) {
    // Test for a non-existing page
    var setEndPoint string = "http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/not"

    response, err := http.Get(setEndPoint)
    if err != nil && response == nil {
       log.Fatalf("Error sending request to API endpoint. %+v", err)
    } else {
        // Close the connection to reuse it
        defer response.Body.Close()

        if response.StatusCode != 404 {
           log.Fatalf("Request was sent but status code is not 404, StatusCode ", response.StatusCode)
        } else {
            body, err := ioutil.ReadAll(response.Body)
            if err != nil {
                log.Fatalf("Couldn't parse response body. %+v", err)
            }
            // Body should be "404 page not found"
	    if string(body) != "404 page not found\n" {
	       log.Fatalf("Response Body: ", string(body))
	    } else {
	      log.Printf("Response Body: %s", string(body))
	    }
        }
    }
}