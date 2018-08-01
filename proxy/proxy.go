package proxy

import (
	"strconv"
	"net/http"
	"github.com/sirupsen/logrus"
	"time"
	"fmt"
	"bytes"
	"io/ioutil"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/store/memory"
	"github.com/ulule/limiter/drivers/middleware/stdlib"
	"io"
	"regexp"
	"strings"
	"github.com/golang/groupcache/lru"
)

type Proxy struct {
	Config ProxyConfig
	Log *logrus.Logger
	whitelistedR []*regexp.Regexp
	blacklsitedR []*regexp.Regexp
	dontCacheR []*regexp.Regexp
	cache *lru.Cache
}

func NewProxy(config ProxyConfig,log *logrus.Logger) *Proxy {
	p := Proxy{Config: config,Log: log}
	return &p
}


func (this *Proxy) Start(){
	this.Log.Info("Starting Proxy!")
	this.Log.Info("Starting with config:", this.Config)
	this.Log.Info("Listening for Connections on Port " + strconv.Itoa(this.Config.ServerPort))
	this.startServer()
}


func (this *Proxy) startServer(){

	setupRegexp(this)
	this.cache = lru.New(this.Config.CacheMaxItems)

	srv := http.Server{
		Addr:         ":" + strconv.Itoa(this.Config.ServerPort),
		ReadTimeout:  time.Duration(this.Config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(this.Config.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(this.Config.IdleTimeout) * time.Second,
	}

	rate := limiter.Rate{
		Period: time.Duration(this.Config.RateLimitPeriod) * time.Second,
		Limit:  this.Config.RateLimitCount,
	}
	store := memory.NewStore()
	middleware := stdlib.NewMiddleware(limiter.New(store, rate), stdlib.WithForwardHeader(true))

	handlerfunc := func(w http.ResponseWriter, req *http.Request) {
		logRequest(this, req)

		tezresponse := []byte(string("Call blacklisted"))

		if this.isAllowed(req.URL.Path) {

			if req.Method == "GET" && this.isCacheable(req.URL.Path){
				fmt.Println("is cacheable")
				if val, ok := this.cache.Get(req.URL.Path); ok {
					tezresponse = val.([]byte)
					fmt.Println("from cache")
				} else {
					tezresponse = this.GetResponse(req)
					this.cache.Add(req.URL.Path,tezresponse)
				}

			} else {
				tezresponse = this.GetResponse(req)
			}

		}

		optionsHeaders(w)
		fmt.Fprint(w, string(tezresponse))
	}
	http.Handle("/", middleware.Handler(http.HandlerFunc(handlerfunc)))
	srv.ListenAndServe()
}

func (this *Proxy) GetResponse(req *http.Request) []byte {
	tezresponse := []byte(string("Method not allowed"))
	if req.Method == "GET" {
		tezresponse = this.GetTezosResponse(req.URL.Path, "")
	}
	if req.Method == "POST" {
		tezresponse = this.PostTezosResponse(req.URL.Path, req.Body)
	}
	return tezresponse
}

func setupRegexp(this *Proxy) {
	for _, s := range this.Config.Blocked {
		regex, err := regexp.Compile(s)
		if err != nil {
			this.Log.Error("Cant compile Regexp: ", s)
		} else {
			this.blacklsitedR = append(this.blacklsitedR, regex)
		}
	}
	for _, s := range this.Config.Methods {
		regex, err := regexp.Compile(s)
		if err != nil {
			this.Log.Error("Cant compile Regexp: ", s)
		} else {
			this.whitelistedR = append(this.whitelistedR, regex)
		}
	}
	for _, s := range this.Config.DontCache {
		regex, err := regexp.Compile(s)
		if err != nil {
			this.Log.Error("Cant compile Regexp: ", s)
		} else {
			this.dontCacheR = append(this.dontCacheR, regex)
		}
	}
}



func (this *Proxy) isAllowed(url string) bool {
	ret := true
	urls := strings.Split(url,"?")
	url = "/" + strings.Trim(urls[0], "/")
	for _,wl := range this.whitelistedR {
		if wl.Match([]byte(url)) {
			for _, bl := range this.blacklsitedR {
				if bl.Match([]byte(url)) {
					ret = false
					break
				}
			}
			break
		}
	}
	return ret
}

func (this *Proxy) isCacheable(url string) bool {
	ret := true
	for _,wl := range this.dontCacheR {
		if wl.Match([]byte(url)) {
			ret = false
		}
	}
	return ret
}


func (this *Proxy) PostTezosResponse(url string, body io.ReadCloser) []byte {
	url = "http://" + this.Config.TezosHost + ":" + strconv.Itoa(this.Config.TezosPort) + url
	bodyb, err := ioutil.ReadAll(body)
	buffer := bytes.NewBuffer(bodyb)
	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		this.Log.Error("Error sending POST to Tezos", err)
	}

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	var b []byte
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		this.Log.Error("Error getting Response from the tezos Node: ", err)
	}
	resp.Body.Close()
	return b
}

func (this *Proxy) GetTezosResponse(url, args string) []byte {
	url = "http://" + this.Config.TezosHost + ":" + strconv.Itoa(this.Config.TezosPort) + url
	var jsonStr = []byte(args)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		this.Log.Error("Error sending GET to Tezos", err)
	}

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	var b []byte
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		this.Log.Error("Error getting Response from the tezos Node: ", err)
	}
	resp.Body.Close()
	return b
}

func logRequest(this *Proxy, req *http.Request) {
	this.Log.WithFields(logrus.Fields{
		"remote_addr": req.RemoteAddr,
		"user_agent":  req.UserAgent(),
	}).Info("Incoming Request")
}

func optionsHeaders(w http.ResponseWriter) {
	w.Header().Set("Allow", "OPTIONS, POST")
	w.Header().Set("Accept", "application/json")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Depth, User-Agent, X-File-Size, X-Requested-With, If-Modified-Since, X-File-Name, Cache-Control")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Content-Type", "application/json")
}