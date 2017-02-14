package main

import (
	"fmt"
	"github.com/jared-d-smith/psl/salestax-srv/lrucache"
	"math/rand"
	"strconv"
	"time"
)

const CACHE_SIZE = 50000
const ATTEMPTS = 10000

func main() {

	c := lrucache.New(CACHE_SIZE)

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < ATTEMPTS; i++ {
		strId := rand.Intn(CACHE_SIZE * 2)
		s := strconv.Itoa(strId)
		c.FastRateLookup(s, sales_tax_lookup)
	}

	fmt.Println(c)
}

// Fake slow lookup routine. The street addresses are stringify'd random numbers
// from [0, CACHE*2]. This routine sleeps for 10ms before returning.
func sales_tax_lookup(key string) (float64, error) {
	val, _ := strconv.ParseInt(key, 10, 64)
	fval := float64(val) * 1.238712
	time.Sleep(10 * time.Millisecond)
	return fval, nil
}
