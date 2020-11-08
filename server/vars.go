package main

import (
	"net/http/httputil"
	"net/url"
	"os"
)

const (
	dateFormat = "Mon Jan _2 15:04:05 2006"
	serverList = "http://127.0.0.1:80,http://127.0.0.1:8080,http://127.0.0.1:9090"
)

//Message in packet
type Message struct {
	Price    int
	Quantity int
	Amount   int
	Object   int
	Method   int
}

//Server s
type Server struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
}

// ServerPool s
type ServerPool struct {
	servers []*Server
	current int
}

var (
	serverPool ServerPool
	counter    int
	limit      int
	port       int
	next       *Server
	file       *os.File = nil
)
