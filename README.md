In this example we're adding Aerospike CacheDB to a URL shortener application

The application is a simple URL shortener. It turns a URL into a hash and returns it in form of a shortened URL.

This application is written in Go. 
It stores its state in an external SQL database and we will add Aerospike CacheDB based caching to the application to improve its performance.

Although this example application is small, it alread benefits from adding Aerospike as a cache layer, with the application data stored in another database.

We implement a Look-Aside caching. Every GET request will issue a cache lookup, if the lookup misses, the data will be retrieven from the SQL storage.

# Go Application Structure

`main.go` - Application defintion
`handlers.go` - HTTP handler funcs
`storage.go` - DB persistence

We will be adding `aerospike.go` which contains the aerospike client setup.

handlers.go will be modified to look up data from Aerospike first, and in case of a miss to populate the data there after requesting it from the application database.

In the end, we will run a performance benchmark before and after the addition.

# Setup The Environment

For the purpose of running this demo you will need the Aerospike Enterprise Docker image for CacheDB running, with the port mapping 3000-3003:3000-3003:

`$ docker run -tid --name aero-ee -p 3000-3003:3000-3003 aerospike-cachedb:latest`

And a Docker image of Postgres with the port mapping 5432:5432:

`$ docker run --name postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres`

# Adding The Cache To The Go Application:

We need to initialize the Aerospike client. The Aerospike-go package provides all necessary functions to communicate and exchange data with a running Aerospike instance.

We import the package:
```
import (
	as "github.com/aerospike/aerospike-client-go"
)
```

```
type Aerospike struct {
	client    *as.Client
	namespace string
}
```

```
const (
	set = "cache"
	ttl = 60
)

func NewAerospike(host string, port int, namespace string) *Aerospike {
	c, err := as.NewClient(host, port)

	if err != nil {
		panic(err)
	}

	return &Aerospike{
		client:    c,
		namespace: namespace,
	}
}
```

```
func (a *Aerospike) Get(hash string) string {
	key, err := as.NewKey(a.namespace, set, hash)

	if err != nil {
		panic(err)
	}

	bin := as.NewBin(hash, nil)
	record, _ := a.client.Get(nil, key, bin.Name)

	if record == nil {
		return ""
	}

	received := record.Bins[bin.Name]

	if received == nil {
		return ""
	} else {
		return fmt.Sprintf("%v", received)
	}
}
```

```
func (a *Aerospike) Add(hash string, val string) {
	key, _ := as.NewKey(a.namespace, set, hash)
	bin := as.NewBin(hash, val)
	wp := as.NewWritePolicy(0, ttl)

	a.client.PutBins(wp, key, bin)
}
```


# Benchmarks Using [wrk](https://github.com/wg/wrk) on a local environment (MacBook Pro 2015)

# WITH LOG NO CACHE
➜  ~ wrk http://localhost:4000/0ea9a5
Running 10s test @ http://localhost:4000/0ea9a5
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    21.40ms   30.79ms 400.18ms   87.79%
    Req/Sec   427.67    146.56   760.00     74.75%
  8456 requests in 10.02s, 1.76MB read
Requests/sec:    843.83
Transfer/sec:    179.64KB

## WITH LOG WITH CACHE
➜  ~ wrk http://localhost:4000/0ea9a5
Running 10s test @ http://localhost:4000/0ea9a5
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    85.58ms  219.80ms   1.01s    88.55%
    Req/Sec     1.65k   368.21     2.17k    82.32%
  27400 requests in 10.06s, 5.70MB read
Requests/sec:   2722.67
Transfer/sec:    579.63KB

Improvement in Requests/sec: x3.2

## NO LOG NO CACHE
➜  ~ wrk http://localhost:4000/0ea9a5
Running 10s test @ http://localhost:4000/0ea9a5
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    14.52ms   18.98ms 143.68ms   86.12%
    Req/Sec   599.49    106.22     0.88k    68.50%
  11940 requests in 10.01s, 2.48MB read
Requests/sec:   1193.29
Transfer/sec:    254.04KB

##  NO LOG WITH CACHE
➜  ~ wrk http://localhost:4000/0ea9a5
Running 10s test @ http://localhost:4000/0ea9a5
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    48.24ms  169.15ms   1.01s    92.88%
    Req/Sec     1.99k   291.48     2.56k    85.71%
  36150 requests in 10.03s, 7.52MB read
Requests/sec:   3605.73
Transfer/sec:    767.63KB

Improvement in Requests/sec: x3