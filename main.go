package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Data struct {
	Apis []Api `json:"data"`
}

type Api struct {
	UpstreamUrl      string `json:"upstream_url"`
	StripRequestPath bool   `json:"strip_request_path"`
	Id               string `json:"id"`
	CreatedAt        int    `json:"created_at"`
	PreserveHost     bool   `json:"preserve_host"`
	Name             string `json:"name"`
	RequestHost      string `json:"request_host"`
}

var host *string

func main() {

	host = flag.String("host", "localhost", "Host address for the kong admin")
	flag.Parse()

	log.Println("Connecting to " + *host + ":8001")

	go func() {
		for range time.Tick(time.Second * 5) {
			resp, err := http.Get("http://" + *host + ":8001/apis")
			defer resp.Body.Close()

			log.Println(resp.StatusCode)

			if err != nil {
				log.Fatal(err)
			}

			decoder := json.NewDecoder(resp.Body)

			var data Data
			err = decoder.Decode(&data)

			if err != nil {
				log.Fatal(err)
			}

			for i := 0; i < len(data.Apis); i++ {
				log.Println(data.Apis[i])
				go Check(data.Apis[i].UpstreamUrl, data.Apis[i].Name)
			}
		}
	}()

	time.Sleep(time.Second * 20)
}

func Check(upstream, name string) {
	resp, err := http.Get(upstream)
	defer resp.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Println(upstream + " = " + strconv.Itoa(resp.StatusCode))

	if resp.StatusCode != 200 {

		// Deregister
		status := Deregister(name)
		log.Println(status)
	}
}

func Deregister(name string) int {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", "http://"+*host+":8001/apis/"+name, nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return resp.StatusCode
}
