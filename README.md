# Kong Beat

Kong Beat is a service listener and health checker for Kong. Very much a work in progress. Please feel free to contribute.

## Features
- Checks status of each service registered in Kong, re-registers services which have stopped.
- Pre-configures Kong services based on given configuration (@TODO)
- Works standalone or in Docke

## How-to build
- Build the binary `$ make`

## How-to run 

### Docker
```
docker run -it --link kong:kong ewanvalentine/kongbeat:latest \ 
       -host=kong \
       -admin-port=8001 \
       -proxy-port=80 \
       -pulse=10 
```

### Standalone 
```
./KongBeat -host=kong \
           -admin-port=8001 \
           -proxy-port=80 \
           -pulse=10 
```

### Docker Compose
```
kong-beat: 
  image: ewanvalentine/kongbeat:latest
  restart: always
  depends_on:
    - kong
    - kong-database
  entrypoint:
    - ./KongBeat
    - -admin-port=8001
    - -proxy-port=80
    - -pulse=5
    - -host=kong
  volumes:
    - /var/run/docker.sock:/var/run/docker.sock

myservice:
  image: myimage
  ports: 
    - "1000:1000"
  environment:
    KONG_UPSTREAM_URL: http://myservice:1000
    KONG_NAME: myservice 
    KONG_HOST: myservce.myhost.com
```

### Idea's 
- Attempt to resuscitate deceased containers using the Docker API?

### Author - Ewan Valentine ewan.valentine89@gmail.com
