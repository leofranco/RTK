# RTK
Small test project for RTK

Using GoLang write a small program which will accomplish the following: 

1. Handle HTTP requests from a web browser asynchronously (concurrently)
2. Respond to two http endpoints with URLs  /get  and /set
3. the /set endpoint will connect to a redis server and create a hashmap with the client's source IP as the key and date + client's user agent as the field/value pair
4. the /get endpoint will connect to a redis server and use the IP url parameter to lookup and return a list of all dates and user agents. The response should be in JSON format.
5. Write test cases for your endpoints
6. Using Siege run some performance tests against your new endpoint and see if you can break your own code. Capture your results. 

## General Comments

The project was not difficult by itself but I had to spent a lot of time catching up with all the technologies involved. I worked on it on Monday and Tuesday (while keeping up with my work) and although there are more improvements I could make and more things I would like to learn I think I got decently far considering the time constraints. A particular area where I would like to learn more is about how to test these kind of applications.

The small program is still running on an aws machine with the following endpoints:

* http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set
* http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/get

with an example using the internal IP of:

* http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/get?IP=10.157.143.26

Please see the file https://github.com/leofranco/RTK/blob/master/ServerRTK3.go for the source code.


## Future Work

At this moment I feel I need more in order to make more progress. The best way to start would be to have real usage patterns. Without them there are only a couple of things that could help in general terms.

* One of them is having a cache as we discussed in Horgen. The thing would be to find a good balance between the speed it saves versus the accuracy. We could re-get the results for a given IP if the data is older than X seconds or we could mark it as obsolete if there has been a new SET from that IP. Also, since a SET only appends data we could add it in the golang logic without going to reds. I just did a quick google for the cache and https://github.com/patrickmn/go-cache would be something easy to try to improve performance.

* I implemented an additional endpoint called set_test that would be useful for benchmarking. We can specify an IP in the same way we do it for the get endpoint. Having a set of benchmarks would be very useful to measure impact of changes in the application (when removing the encoding/decoding from son for instance).

#### There are several improvements that would make sense only after everything else has been optimised.

* Checkout the parameter for redis pools and optimise that.
* Avoid as many encode/decode operations from JSON as possible (we could store the end result in the reds database and save the conversion).
* Check all the String operations in the program (concatenations are expensive). The SplitHostPort for instance is probably not needed.
* See if using all the CPUs hurts instead of helping (too many context switches).
* Try a more optimised version of net/http like https://github.com/valyala/fasthttp
* See if the other redis libraries are better/faster. I just found out go-redis has a benchmark comparison vs redigo.


## Journal
##### I wrote many comments as I was working on it, some are not very useful but they can help understand my train of thought.

###### Create a new VM in aws.amazon 

Make things simple and start with the default Ubuntu (Ubuntu Server 14.04 LTS (HVM), SSD Volume Type - ami-2d39803a)

For the redis database the biggest impact will probably be memory size (note to self: check if golang is inherently concurrently and will be able to use multiple threads, if it is it makes sense to get more cpu’s).

There is a machine in amazon with 1952GB of memory and 128 cores (it is only $13 an hour too). I would need a whole cluster to bring that one down during tests. For the moment I will be reasonable and I will take the second cheapest machine from the memory optimised options. Four cores, 30GB in memory and a 80GB SSD drive for $0.333 an hour.

Launch the machine and setup pass-less login with one of the existing certificates. Remember to modify the ssh config site to include ubuntu as the default user for that address (I will leave the ubuntu user in that testing machine).

###### Install Redis 
The default configuration is on port 6379 (remember to open it up on AWS if needed… and it looks like there is no auth, be careful)
We can use redis-cli to test the installation 

###### Install Go 
$ go version
go version go1.2.1 linux/amd64

###### Install go redis
go get github.com/garyburd/redigo/redis

###### Install Siege
$ curl -O http://download.joedog.org/siege/siege-latest.tar.gz

Siege installed in my laptop in Split, Croatia (264.96Mbps Down - 238.89Mbps Up).

Both set and siege where launched at the same time from my laptop

```
siege -v -r 200 -c 255 -b http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set
siege -v -r 200 -c 255 -b http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/get
```

At that moment the server broke. It looks like it cannot have so many open connections. Redis for sure was affected seeing the warning *dial tcp :6379: too many open files*

Something interesting is that only one core was working at 100%, the whole system was only at 25% (4 cores) which makes me wonder how is the concurrency handled by gaoling and redis.

```
2016/09/26 19:55:55 http: panic serving 212.91.120.224:1719: dial tcp :6379: too many open files
goroutine 52252 [running]:
net/http.func·009()
           /usr/lib/go/src/pkg/net/http/server.go:1093 +0xae
runtime.panic(0x649a80, 0xc336448b00)
           /usr/lib/go/src/pkg/runtime/panic.c:248 +0x106
main.HandlerSet(0x7fba61795500, 0xc215ba0f00, 0xc221e61680)
           /home/ubuntu/RTK/main.go:54 +0x1a5
…
2016/09/26 19:55:57 http: Accept error: accept tcp [::]:8080: too many open files; retrying in 5ms
```

I tried again only with the set:
```
$ siege -v -r 200 -c 255 -b http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set
```

siege aborted due to excessive socket failure; you
can change the failure threshold in $HOME/.siegerc

Transactions:                 28232 hits
Availability:                 95.67 %
Elapsed time:                 51.41 secs
Data transferred:             3.74 MB
Response time:                0.44 secs
Transaction rate:             549.15 trans/sec
Throughput:                   0.07 MB/sec
Concurrency:                  240.73
Successful transactions:      28232
Failed transactions:               1278
Longest transaction:               3.11
Shortest transaction:              0.29

and the go program broker again. I will have to implement a pool for the connections to redis. That is the first issue to solve. The second thing is to find out why there is only one core being used.

The second try is on file ServerRTK2.go

If we try the siege again:
```
$ siege -v -r 200 -c 255 -b http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set
```
Transactions:                51000 hits
Availability:                100.00 %
Elapsed time:                95.03 secs
Data transferred:            6.76 MB
Response time:               0.44 secs
Transaction rat              536.67 trans/sec
Throughput:                  0.07 MB/sec
Concurrency:                 236.67
Successful transactions:     51000
Failed transactions:                  0
Longest transaction:               3.93
Shortest transaction:              0.29

Not too bad, at least the program survived and the cpu usage didn’t go above 20% of one core

I will clean the database and try the siege again with set and get (at the same time):
```
$ siege -v -r 200 -c 255 -b http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set
```
Transactions:                51000 hits
Availability:                100.00 %
Elapsed time:                102.71 secs
Data transferred:            6.76 MB
Response time:               0.46 secs
Transaction rate:            496.54 trans/sec
Throughput:                  0.07 MB/sec
Concurrency:                 229.21
Successful transactions:       51000
Failed transactions:                  0
Longest transaction:               7.57
Shortest transaction:              0.29

On the get side the load was stable at 6% of one core but the memory increased substantially (>10GB)

I am wondering why only one core is used although the requests are handled concurrently. I tried another tool called ab (although that one is single-threaded by itself):
```
$ ab -c 500 -n 500 http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set
This is ApacheBench, Version 2.3 <$Revision: 1528965 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking ec2-107-22-107-4.compute-1.amazonaws.com (be patient)
Completed 100 requests
Completed 200 requests
Completed 300 requests
Completed 400 requests
Completed 500 requests
Finished 500 requests


Server Software:
Server Hostname:        ec2-107-22-107-4.compute-1.amazonaws.com
Server Port:            8080

Document Path:          /set
Document Length:        102 bytes

Concurrency Level:      500
Time taken for tests:   0.211 seconds
Complete requests:      500
Failed requests:        53
   (Connect: 0, Receive: 0, Length: 53, Exceptions: 0)
Total transferred:      109939 bytes
HTML transferred:       50940 bytes
Requests per second:    2367.19 [#/sec] (mean)
Time per request:       211.221 [ms] (mean)
Time per request:       0.422 [ms] (mean, across all concurrent requests)
Transfer rate:          508.29 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    2   2.7      0       7
Processing:     3   30  31.2     14     203
Waiting:        3   30  31.2     14     203
Total:          3   32  33.7     14     209

Percentage of the requests served within a certain time (ms)
  50%     14
  66%     16
  75%     81
  80%     82
  90%     86
  95%     87
  98%     87
  99%     87
 100%    209 (longest request)
```

* So, I found out that net/http already starts a goroutine per request. The problem is that you have to tell go that it is allowed to use more than one CPU. The solution looks simple enough:

```
runtime.GOMAXPROCS(runtime.NumCPU())
```
I tried again after that change (to use 4 cores) and run the 4 commands at the same time:

```
$ siege -v -r 500 -c 255 -b http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set
$ siege -v -r 500 -c 255 -b http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set
$ siege -v -r 10 -c 255 -b http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/get
$ siege -v -r 50 -c 255 -b http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/get
```

but at this point the bottleneck is my laptop because it died.

Trying again a simple siege:
```
$ siege -v -r 500 -c 255 -b http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set
```
Transactions:                127500 hits
Availability:                100.00 %
Elapsed time:                237.01 secs
Data transferred:            16.89 MB
Response time:               0.45 secs
Transaction rate:            537.95 trans/sec
Throughput:                  0.07 MB/sec
Concurrency:                 242.34
Successful transactions:     127500
Failed transactions:                  0
Longest transaction:              10.76
Shortest transaction:              0.29

I am going to start another was machine but this one needs more cpus and less memory. I am going to take a c4.2xlarge with 8 cores at $0.419 per Hour.

I will try the same siege above but started 8 times at the same time:

==> nohup_t1.out <==
siege aborted due to excessive socket failure; you
can change the failure threshold in $HOME/.siegerc

Transactions:                 1976 hits
Availability:                 61.69 %
Elapsed time:                 45.47 secs
Data transferred:             0.26 MB
Response time:                0.93 secs
Transaction rate:             43.46 trans/sec
Throughput:                   0.01 MB/sec
Concurrency:                  40.23
Successful transactions:      1976
Failed transactions:               1227
Longest transaction:               5.09
Shortest transaction:              0.00

All the outputs are similar which means we are losing transactions but the cpu in the server is still not being used.

Now a test with ab:

```
$ ab -n 1000 -c 1000 http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set
This is ApacheBench, Version 2.3 <$Revision: 1528965 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking ec2-107-22-107-4.compute-1.amazonaws.com (be patient)
Completed 100 requests
Completed 200 requests
Completed 300 requests
Completed 400 requests
Completed 500 requests
Completed 600 requests
Completed 700 requests
Completed 800 requests
Completed 900 requests
Completed 1000 requests
Finished 1000 requests


Server Software:
Server Hostname:        ec2-107-22-107-4.compute-1.amazonaws.com
Server Port:            8080

Document Path:          /set
Document Length:        102 bytes

Concurrency Level:      1000
Time taken for tests:   0.252 seconds
Complete requests:      1000
Failed requests:        98
   (Connect: 0, Receive: 0, Length: 98, Exceptions: 0)
Total transferred:      219893 bytes
HTML transferred:       101894 bytes
Requests per second:    3970.65 [#/sec] (mean)
Time per request:       251.848 [ms] (mean)
Time per request:       0.252 [ms] (mean, across all concurrent requests)
Transfer rate:          852.65 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        2   11   6.6      9      22
Processing:     4   36  36.5     29     206
Waiting:        3   36  36.5     29     206
Total:         15   47  35.2     43     223

Percentage of the requests served within a certain time (ms)
  50%     43
  66%     48
  75%     50
  80%     52
  90%     59
  95%     67
  98%    208
  99%    208
 100%    223 (longest request)
```

Note: just discovered the testonborrow from the redigo example sends a ping request when you ask for the resource. I am removing it now.


This is the last test I ran at 8pm on Tuesday:
```
siege -v -r 100 -c 255 -b http://ec2-107-22-107-4.compute-1.amazonaws.com:8080/set
```
Transactions:  		       25500 hits
Availability:  		      100.00 %
Elapsed time:  		       97.67 secs
Data transferred:      	        2.36 MB
Response time: 		        0.82 secs
Transaction rate:      	      261.08 trans/sec
Throughput:    		        0.02 MB/sec
Concurrency:   		      214.63
Successful transactions:                25500
Failed transactions:   	           0
Longest transaction:   	        5.13
Shortest transaction:  	        0.00
