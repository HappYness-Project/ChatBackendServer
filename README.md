# ChatServerDev


### Build Dockerfile
build the Dockerfile into an image with image name 'my-app'
```
docker build -t my-app:latest .
```
### Start a new container from a Docker image
Runs the container in interactive mode with a pseudo-TTY on port 4545
```
docker run -it -p 4545:4545 --name chat-server my-app 
```