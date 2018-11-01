package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hoenigmann/redis-proxy/proxy"
)

type options struct {
	redisHost       string
	redisPort       string
	cacheExpiryTime string
	capacity        int
}

var lg *log.Logger

func init() {
	lg = log.New(os.Stdout, "", log.LstdFlags)
}

func main() {
	opt := getOptions()
	fmt.Println(opt)
	proxy := proxy.New(opt.redisHost, opt.redisPort, fmtExpiry(opt.cacheExpiryTime), opt.capacity)
	handler := NewHttpHandler(proxy)

	log.Fatal(http.ListenAndServe("localhost:80", handler))
}

func fmtExpiry(expiryTime string) time.Duration {
	expiry := expiryTime + "m"
	dur, err := time.ParseDuration(expiry)
	if err != nil {
		fmt.Println("Not a valid global expiration format: " + expiryTime + " (integer in minutes)")
		panic(err)
	}
	return dur
}

type HttpHandler struct {
	proxy *proxy.Proxy
}

func NewHttpHandler(proxy *proxy.Proxy) *HttpHandler {
	return &HttpHandler{proxy}
}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.processRequest(w, req)
}

func (h *HttpHandler) processRequest(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if req.Method != "GET" {
		return
	}
	path := req.RequestURI
	key := strings.TrimPrefix(path, "/")

	lg.Println("Retrieving key from Proxy Cache: ", key)
	if result, ok := h.proxy.Get(key); ok {

		w.Header().Set("Content-Type", "text/plain")
		_, err := fmt.Fprintf(w, result)
		if err != nil {
			panic(err)
		}
	} else {
		w.WriteHeader(404)
	}
}

func getOptions() *options {
	opt := new(options)
	flag.StringVar(&opt.redisHost, "h", "localhost", "The redis host address")
	flag.StringVar(&opt.redisPort, "p", "6379", "The redis port to connect to")
	flag.StringVar(&opt.cacheExpiryTime, "e", "10", "global time until a entry expires in cache (in seconds)")
	flag.IntVar(&opt.capacity, "c", 10000, "The size of the cache in terms of the number of keys")

	flag.Parse()

	return opt
}
