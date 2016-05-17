default: build

build:
	GOOS=linux GOARCH=amd64 go build
	docker build -t theladbiblegroup/kongbeat:latest . 

run: 
	docker run -it theladbiblegroup/kongbeat:latest -host=65twenty.com
