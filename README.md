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
       -host=localhost \
       -admin-port=8001 \
       -proxy-port=80 \
       -pulse=10 
```

### Standalone 
```
./KongBeat -host=localhost \
           -admin-port=8001 \
           -proxy-port=80 \
           -pulse=10 
```

### Idea's 
- Pre-configure services, YAML definition file? 
- Attempt to resuscitate deceased containers using the Docker API?

### Author - Ewan Valentine ewan.valentine89@gmail.com
