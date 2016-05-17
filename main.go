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
var pulse *int
var KongProxy *int
var KongAdmin *int

func main() {

	host = flag.String("host", "localhost", "Host address for the kong admin")
	pulse = flag.Int("pulse", 5, "Refresh rate for api checks in seconds")
	KongProxy = flag.Int("proxy-port", 8000, "Proxy port for Kong")
	KongAdmin = flag.Int("admin-port", 8001, "Admin port for Kong")
	flag.Parse()

	log.Println("Connecting to " + *host + ":" + strconv.Itoa(*KongAdmin))

	go func() {
		for range time.Tick(time.Second * time.Duration(*pulse)) {
			resp, err := http.Get("http://" + *host + ":" + strconv.Itoa(*KongAdmin) + "/apis")
			defer resp.Body.Close()

			log.Println("Heartbeat:", resp.StatusCode)

			if resp == nil {
				log.Fatal(err)
			}

			decoder := json.NewDecoder(resp.Body)

			var data Data
			err = decoder.Decode(&data)

			if err != nil {
				log.Fatal(err)
			}

			// Foreach API
			for i := 0; i < len(data.Apis); i++ {

				// Check
				status := Check(
					data.Apis[i].RequestHost+":"+strconv.Itoa(*KongProxy),
					data.Apis[i].Name,
				)

				// If status not 200, de-register service
				if status != 200 {
					go Deregister(data.Apis[i].Name)
				}
			}
		}
	}()

	done := make(chan bool)
	go forever()
	<-done
}

func forever() {
	for {
		time.Sleep(time.Second)
	}
}

// Check - Check a service upstream, return status
func Check(host, name string) int {

	// 2 second timeout, timeout shouldn't be really long
	client := http.Client{
		Timeout: time.Duration(2 * time.Second),
	}
	resp, _ := client.Get(host)

	if resp != nil {
		log.Println("OK:", host+" -- "+strconv.Itoa(resp.StatusCode))
		return resp.StatusCode
	}

	log.Println("Lost:", name)
	return 404
}

// Deregister - Deregister a service
func Deregister(name string) {
	log.Println("De-registering service: " + name)
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", "http://"+*host+":8001/apis/"+name, nil)
	resp, err := client.Do(req)

	if resp != nil {
		log.Println("De-registered service: " + strconv.Itoa(resp.StatusCode))
	}

	log.Println(err)
}
