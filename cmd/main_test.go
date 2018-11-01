package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/mediocregopher/radix.v2/redis"

	"github.com/hoenigmann/redis-proxy/proxy"
)

func TestSystem(t *testing.T) {
	//backing := backing.New("localhost", "6379")

	proxy := proxy.New("localhost", "6379", time.Second*5, 10)
	handler := NewHttpHandler(proxy)
	go http.ListenAndServe(":8080", handler)

	// These are only found at first in backing store (redis).
	testGetValid(t, "http://localhost:8080/a", "b")
	testGetValid(t, "http://localhost:8080/b", "67")
	// Now in proxy cache
	testGetValid(t, "http://localhost:8080/a", "b")
	testGetValid(t, "http://localhost:8080/b", "67")

	// Does not exist.
	testGetNoExist(t, "http://localhost:8080/c")

	testGetNoExist(t, "http://localhost:8080/d")
	// put in backing redis with 1 minute expiry.
	c, _ := redis.Dial("tcp", "localhost:6379")
	resp := c.Cmd("SET", "d", "42", "NX", "EX", "5")
	if resp.Err != nil {
		t.Errorf("%v", resp.Err)
	}
	time.Sleep(time.Second * 1)
	// Within global expiry of cache, should be there.
	testGetValid(t, "http://localhost:8080/d", "42")
	time.Sleep(time.Second * 1)
	// Still within global expiry of cache, should be there.
	testGetValid(t, "http://localhost:8080/d", "42")
	time.Sleep(time.Second * 4)
	// Outside of global expiry of proxy cache (5 seconds), should not longer exist in redis or in proxy cache.
	testGetNoExist(t, "http://localhost:8080/d")

	for i := 0; i < 1000; i++ {
		testGetValid(t, "http://localhost:8080/a", "b")
		testGetValid(t, "http://localhost:8080/b", "67")
	}
}

func testGetValid(t *testing.T, url string, correct string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Errorf("Error on getting a key that was in redis backing %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("%v didnt respond ok: %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		t.Errorf("%v body didn't parse ok", url)
	}

	assertEqual(t, string(b), correct, "")
}

func testGetNoExist(t *testing.T, url string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Errorf("Error on getting a key that was in redis backing %v", err)
	}

	assertEqual(t, resp.StatusCode, 404, "Status code not 404")
}

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}
