# Planer
Webapplikation zur Erstellung eines Messdienerplans


# Install Planer on RaspberryPi as Webservice within Docker

## Setup PI
First we need to setup the Pi with an operating system. Because we run our server inside a Docker Container, I recomend to use an OS that already inclue Docker. I use **Hypriot**. You can find the instruction how to Install Hypriot here:


[Hypriot - Docker on RPI](https://blog.hypriot.com/getting-started-with-docker-on-your-arm-device/)

To connect first time to the pi use ssh

```ssh pirate@LOCAL-IP-OF-PI```

The default PW for the user `pirate` is `hypriot`

## Setup Portainer (optional)
An easy way to manage all Docker container is  portainer.io

*Portainer is a lightweight management UI which allows you to easily manage your Docker host or Swarm cluster.*

## Get Files
First we create a new folder for our Planer Service with:
```
mkdir planer && cd planer
```

Then we download the Dockerfile and the docker-compose.yml from this github repo with:
```
curl --get https://raw.githubusercontent.com/Maekes/planer/master/Dockerfile > Dockerfile
curl --get https://raw.githubusercontent.com/Maekes/planer/master/docker-compose.yml > docker-compose.yml
```

## Start the service
After getting all file we can start our application with
```
docker-compose up
```