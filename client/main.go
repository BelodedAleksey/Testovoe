package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	server     = "http://127.0.0.1:80"
	dateFormat = "Mon Jan _2 15:04:05 2006"
)

//Message s
type Message struct {
	Price    int
	Quantity int
	Amount   int
	Object   int
	Method   int
}

var (
	freqMessage int
	file        *os.File = nil
)

func main() {
	flag.IntVar(&freqMessage, "f", 10, "frequency")
	flag.Parse()
	t := time.NewTicker(time.Millisecond * time.Duration(freqMessage))
	for {
		select {
		case <-t.C:
			sendPacket()
		}
	}
}

//write to log file
func logFile(data string) {
	if file == nil {
		file, _ = os.OpenFile(
			"log client.txt",
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

func sendPacket() {
	numPackets := rand.Intn(5)
	buffer := new(bytes.Buffer)
	packet := []Message{}
	for i := 0; i < numPackets; i++ {
		m := Message{
			Price:    rand.Intn(100),
			Quantity: rand.Intn(100),
			Amount:   rand.Intn(100),
			Object:   rand.Intn(100),
			Method:   rand.Intn(100),
		}
		packet = append(packet, m)
		data := fmt.Sprintf(
			"Packet№ %d\nprice: %d\nquantity: %d\namount: %d\nobject: %d\nmethod: %d\n",
			i, m.Price, m.Quantity, m.Amount, m.Object, m.Method,
		)
		logFile(data)
	}
	json.NewEncoder(buffer).Encode(packet)
	fmt.Println("Length of packet: ", buffer.Len())
	req, err := http.NewRequest("POST", server, buffer)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	client.Timeout = time.Second * 30
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	logFile("response Status: " + resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
