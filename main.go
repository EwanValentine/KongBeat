package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/fsouza/go-dockerclient"
	"github.com/kataras/iris"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Data struct {
	Apis []Api `json:"data"`
}

type Api struct {
	UpstreamUrl      string `json:"upstream_url"`
	StripRequestPath bool   `json:"strip_request_path"`
	PreserveHost     bool   `json:"preserve_host"`
	Name             string `json:"name"`
	RequestHost      string `json:"request_host"`
}

var port *int
var host *string
var pulse *int
var KongProxy *int
var KongAdmin *int

func main() {

	port = flag.Int("port", 4242, "Port Kong runs on")
	host = flag.String("host", "localhost", "Host address for the kong admin")
	pulse = flag.Int("pulse", 5, "Refresh rate for api checks in seconds")
	KongProxy = flag.Int("proxy-port", 8000, "Proxy port for Kong")
	KongAdmin = flag.Int("admin-port", 8001, "Admin port for Kong")
	flag.Parse()

	log.Println("Connecting to " + *host + ":" + strconv.Itoa(*KongAdmin))

	iris.StaticServe("./static", "/")

	iris.Get("/", func(ctx *iris.Context) {
		ctx.Write("Append one of these to browser's address bar:\n/assets/js/jquery-2.1.1.js\n/assets/css/bootstrap.min.css")
	})

	iris.StaticServe("./resources", "/assets")

	log.Println("Running on port", *port)

	log.Println("testing")

	dockerEvents := make(chan *docker.APIEvents)
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)

	if err != nil {
		log.Fatal(err)
	}

	// Does initial docker check
	DockerCheck(dockerEvents, client)

	// Sets event listener on further docker events
	go DockerListen(dockerEvents, client)

	// Listens for changes in Kong api's
	go func() {
		for range time.Tick(time.Second * time.Duration(*pulse)) {
			resp, err := http.Get("http://" + *host + ":" + strconv.Itoa(*KongAdmin) + "/apis")
			// defer resp.Body.Close()

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
					data.Apis[i].UpstreamUrl,
					data.Apis[i].Name,
				)

				log.Println(status)

				// If status not 200, de-register service
				if status != 200 {

					// Alerting will go here
					// go Deregister(data.Apis[i].Name)
				}
			}
		}
	}()

	iris.Listen("0.0.0.0:" + strconv.Itoa(*port))

	// Might not need this what with the web server
	done := make(chan bool)
	go forever()
	<-done
}

func forever() {
	for {
		time.Sleep(time.Second)
	}
}

// GetKongEnvVars - Get Kong environment variable from a Docker container
func GetKongEnvVars(container *docker.Container) Api {

	var api Api

	// Get environment variables, look for KONG_BEAT_*
	// Foreach environment variable
	for _, env := range container.Config.Env {

		// Look for Kong upstream url
		if strings.HasPrefix(env, "KONG_UPSTREAM_URL=") {
			api.UpstreamUrl = env[len("KONG_UPSTREAM_URL")+1:]
		}

		// Look for kong service name
		if strings.HasPrefix(env, "KONG_NAME=") {
			api.Name = env[len("KONG_NAME")+1:]
		}

		if strings.HasPrefix(env, "KONG_HOST=") {
			api.RequestHost = env[len("KONG_HOST")+1:]
		}

		// @todo - do preserve host and other opts
	}

	return api
}

// DockerCheck - Does initial checks for docker containers
func DockerCheck(dockerEvents chan *docker.APIEvents, client *docker.Client) {
	containers, _ := client.ListContainers(docker.ListContainersOptions{All: true})

	for _, containerItem := range containers {
		if container, _ := client.InspectContainer(containerItem.ID); container != nil {
			api := GetKongEnvVars(container)
			if api.Name != "" {
				go Register(api)
			}
		}
	}
}

// DockerListen - Listens for docker `start` events
func DockerListen(dockerEvents chan *docker.APIEvents, client *docker.Client) {

	err := client.AddEventListener(dockerEvents)

	if err != nil {
		log.Fatal(err)
	}

	for event := range dockerEvents {
		if event.Status == "start" {
			if container, _ := client.InspectContainer(event.ID); container != nil {

				api := GetKongEnvVars(container)

				if api.Name != "" {
					go Register(api)
				}
			}
		}
	}
}

// Check - Check a service upstream, return status
func Check(upstream, name string) int {

	// 2 second timeout, timeout shouldn't be really long
	client := http.Client{
		Timeout: time.Duration(2 * time.Second),
	}
	resp, _ := client.Get(upstream)

	if resp != nil {
		log.Println("OK:", upstream+" -- "+strconv.Itoa(resp.StatusCode))
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

// Register - Register a service with Kong
func Register(api Api) {

	log.Println("Registering Service:", api.Name)

	client := &http.Client{}
	data, _ := json.Marshal(api)
	req, err := http.NewRequest("POST", "http://"+*host+":8001/apis", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp != nil {
		if resp.StatusCode == 500 {
			log.Fatal("Failed to register service:", resp.StatusCode)
		}
		log.Println("Successfully registered service:", api.Name)
	} else {
		log.Fatal("Couldn't connect to kong:", err)
	}
}
