package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

// AddServer to the server pool
func (s *ServerPool) AddServer(server *Server) {
	s.servers = append(s.servers, server)
}

//NextServer f
func (s *ServerPool) NextServer() *Server {
	if s.current == len(s.servers)-1 {
		s.current = 0
		return s.servers[0]
	}
	s.current++
	return s.servers[s.current]
}

func handler(w http.ResponseWriter, r *http.Request) {
	//check freq of client's requests
	if counter >= limit {
		next.ReverseProxy.ServeHTTP(w, r)
		return
	}
	counter++

	//Parsing json
	var packet []Message
	err := json.NewDecoder(r.Body).Decode(&packet)
	if err != nil {
		log.Println("Error: \n", err)
		logFile(err.Error())
		return
	}

	//log received packet
	for i, m := range packet {
		data := fmt.Sprintf(
			"Packet№ %d\nprice: %d\nquantity: %d\namount: %d\nobject: %d\nmethod: %d\n",
			i, m.Price, m.Quantity, m.Amount, m.Object, m.Method,
		)
		logFile(data)
	}

	//write response to client
	response := `
	<!DOCTYPE html>
    <html>
    <head>
      <meta charset='utf-8'>
      <meta name='viewport' content='width=device-width'>
      <title>Message</title>
      <style> body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; padding:1em; } </style>
    </head>
	<body>
	Json was parsed!                  
    </body>
    </html>
      `
	w.Write([]byte(response))
	return
}

//reset number of packets every second
func resetCounter() {
	t := time.NewTicker(time.Second)
	for {
		select {
		case <-t.C:
			log.Printf("Port: %d Num of packets: %d\n", port, counter)
			counter = 0
		}
	}
}

//write to log file
func logFile(data string) {
	if file == nil {
		file, _ = os.OpenFile(
			fmt.Sprintf("log %d.txt", port),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
	}

	when := time.Now().UTC().Format(dateFormat)
	data = when + "\n" + data
	if _, err := file.Write([]byte(data)); err != nil {
		log.Println("[LOG] ", err)
	}
}

func main() {
	flag.IntVar(&port, "port", 80, "Port to serve")
	flag.IntVar(&limit, "limit", 100, "limit of messages per sec")
	flag.Parse()

	// parse servers
	tokens := strings.Split(serverList, ",")
	for _, tok := range tokens {
		serverURL, err := url.Parse(tok)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverURL)
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				Renegotiation:      tls.RenegotiateFreelyAsClient,
			},
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 5 * time.Second,
			IdleConnTimeout:       5 * time.Second,
		}
		proxy.Transport = transport
		proxy.ModifyResponse = func(resp *http.Response) error {
			//fix port changing
			resp.Header.Set("Location", serverURL.String())
			return nil
		}
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			data := fmt.Sprintf("[%s] %s\n", serverURL.Host, e.Error())
			log.Println(data)
			logFile(data)
			next = serverPool.NextServer()
		}

		serverPool.AddServer(&Server{
			URL:          serverURL,
			ReverseProxy: proxy,
		})
		log.Printf("Add server: %s\n", serverURL)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(handler),
	}

	next = serverPool.NextServer()

	go resetCounter()

	data := fmt.Sprintf("Server started at %s\n", server.Addr)
	log.Printf(data)
	logFile(data)

	if err := server.ListenAndServe(); err != nil {
		logFile(err.Error())
		log.Fatal(err)
	}
}
