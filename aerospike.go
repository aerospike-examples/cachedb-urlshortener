package main

import (
	"fmt"

	as "github.com/aerospike/aerospike-client-go"
)

type Aerospike struct {
	client    *as.Client
	namespace string
}

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

func (a *Aerospike) Add(hash string, val string) {
	key, _ := as.NewKey(a.namespace, set, hash)
	bin := as.NewBin(hash, val)
	wp := as.NewWritePolicy(0, ttl)

	a.client.PutBins(wp, key, bin)
}
