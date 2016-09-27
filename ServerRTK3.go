package main

// Using GoLang write a small program which will accomplish the following: 
//
// 1) Handle HTTP requests from a web browser asynchronously (concurrently)
// 2) Respond to two http endpoints with URLs  /get  and /set
// 3) the /set endpoint will connect to a redis server and create a 
//    hashmap with the client's source IP as the key and date + client's user 
//    agent as the field/value pair
// 4) the /get endpoint will connect to a redis server and use the IP url 
//    parameter to lookup and return a list of all dates and user agents. 
//    The response should be in JSON format.
// 5) Write test cases for your endpoints
// 6) Using Siege run some performance tests against your new endpoint 
//    and see if you can break your own code. Capture your results. 

// The idea of the exercise is to serve
// multiple requests concurrently.
// As a first try I am letting Go do all
// the work since my understanding is
// that new subruoutines are created when 
// the http.HandleFunc is called.
// The problem is that if that routine takes
// a long time we run the risk of spawning
// and unmanageable number of subroutines
// (and the same can be said about the redis
// connection that is inside the handlers)

// This is try Number 2
// I am implementing a connection pool
// for the redis database since the program
// breaks when  we have too many requests
// Google Redigo pool

import (
    "fmt"
    "time"
    "net"
    "net/http"
    "github.com/garyburd/redigo/redis"
    "encoding/json"
    "runtime"
)

func newPool() *redis.Pool {
    return &redis.Pool{
        MaxIdle: 64,
	MaxActive: 1024,
	Wait: true,
        IdleTimeout: 1 * time.Second,
        Dial: func () (redis.Conn, error) {
            c, err := redis.Dial("tcp", ":6379")
            if err != nil {
                panic(err.Error())
            }
            return c, err
        },
    }
}
var pool = newPool()

func HandlerSet(w http.ResponseWriter, r *http.Request) {
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        http.Error(w, `Invalid Address`, http.StatusBadRequest) 
        return
    }    
    var user_agent string = r.UserAgent()

    // time parsing is slow.
    // also be careful with time zones
    time_utc := time.Now().UTC() 

    // We don't have a connection per request
    // now, just a fixed number in the pool
    redis_conn := pool.Get()
    defer redis_conn.Close()

    // redis> HSET myhash field1 "Hello"
    // redis> HGET myhash field1

    //redis_conn.Do("HSET", "user:"+ip, "ip", ip)
    //redis_conn.Do("HSET", "user:"+ip, "user_agent", user_agent)
    //redis_conn.Do("HSET", "user:"+ip, "time_utc", time_utc)

    // instead of having field names we are
    // using the date/useragent as a filed/value pair
    redis_conn.Do("HSET", "user:"+ip, time_utc, user_agent)
    
    fmt.Fprintf(w, "IP: %s\n", ip)
    fmt.Fprintf(w, "UserAgent: %s\n", user_agent)
    fmt.Fprintf(w, "TimeStamp (UTC): %s\n", time_utc)
}

func HandlerGet(w http.ResponseWriter, r *http.Request) {
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        http.Error(w, `Invalid Address`, http.StatusBadRequest)
        return
    }
    
    // We don't    have a connection per request
    // now, just a fixed number    in the pool
    redis_conn := pool.Get()
    defer redis_conn.Close()

    // Here we have to be careful
    // There is the option of 
    // specifying the URL for which
    // we want to extract the data
    // This whole block is way too hackish
    // 2do remove all the hard burned stuff
    ipreq := r.URL.Query()["IP"]
    if ipreq != nil {
       	   // fmt.Fprintf(w,"Replacing IP %s, by IP %s", ip, ipreq[0])
           ip=ipreq[0]
    }

    // Get all the field/value pairs
    // for this IP address
    res_list, err := redis.StringMap(redis_conn.Do("HGETALL", "user:"+ip))
    if err != nil {
           panic(err)
    }
    
    // We are treating the redis results as
    // string,string map. Since the structure
    // is already taken care of we can encode it
    // in json
    encoded, _ := json.Marshal(res_list)
    encoded_string := string(encoded)
    fmt.Fprintf(w,encoded_string)
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    fmt.Printf("Number of cores enabled: %d\n", runtime.GOMAXPROCS(runtime.NumCPU()))

    http.HandleFunc("/set", HandlerSet)
    http.HandleFunc("/get", HandlerGet)
    http.ListenAndServe(":8080", nil)
}
