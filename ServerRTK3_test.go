package main

import (
    "time"
    "net/http"
    "testing"
    "log"
    "io/ioutil"
    "math/rand"
    "strings"
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

// Do a harder test
// Call the set operation 5 times for a particular IP
// Then call the get operation and test it has 5 fields
func TestSetTest(t *testing.T) {
    s2 := rand.NewSource(time.Now().UnixNano())
    r2 := rand.New(s2)
    var rint int = r2.Intn(100000) 
    var setTestEndPoint string = "http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set_test?IP=" + string(rint)

    for i := 0; i < 5; i++ {
        response, err := http.Get(setTestEndPoint)
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
    } //for


    var getEndPoint string = "http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/get?IP=" + string(rint)

    response, err := http.Get(getEndPoint)
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
	    // If we are here the body should have 5 fields
            

            // Split on comma.
            tokens := strings.Split(string(body), ",")

            // Length is 3.
            if len(tokens) != 5 {
                log.Fatalf("Wrong number of fields for 5 GETs : %d", len(tokens))
	    } else {
    	        log.Println("Response Body:", string(body))
	    }
        }
    }
}