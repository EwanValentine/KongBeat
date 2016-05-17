# Kong Beat

Kong Beat is a service listener and health checker for Kong. Very much a work in progress. Please feel free to contribute.

## Features
- Checks status of each service registered in Kong, re-registers services which have stopped.
- Pre-configures Kong services based on given configuration (@TODO)
- Works standalone or in Docker

## How-to build
- Build the binary `$ make`

## How-to run 

### Docker
- `docker run -it --link kong:kong theladbiblegroup/kongbeat:latest -host=localhost`

### Standalone 
- `./KongBeat -host=localhost`



