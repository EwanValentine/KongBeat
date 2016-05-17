default: build

build:
	GOOS=linux GOARCH=amd64 go build
	docker build -t ewanvalentine/kongbeat:latest . 

run: 
	docker run -it ewanvalentine/kongbeat:latest -host=65twenty.com
