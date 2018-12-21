#!/bin/sh

# Replace with the name of your executable
EXEC="go-mongodb-service"

if [ "$1" = "start" ]
then 
    echo -e "\nStarting service $2"
    cd $2
    ./$EXEC &>/dev/null &disown 
    echo -e "Service $2 started"
fi

if [ "$1" = "stop" ]
then 
    PID=$(ps -ef | grep $EXEC | grep -v grep | awk '{print $2}')

    kill $PID
    echo -e "Service stopped"
fi

if [ "$1" = "build" ]
then
  echo -e "\nBuilding application $2"
  cd $2
  go build .
  echo -e "Application $2 built"
fi

if [ "$1" = "test" ]
then
  # local testing with docker images
  docker run -d --rm --name mongodb-service -p 27017:27017 -e MONGODB_USER=pubcode -e MONGODB_PASSWORD=pubcode -e MONGODB_DATABASE=pubcodes -e MONGODB_ADMIN_PASSWORD=admin -v /home/lzuccarelli/Data:/var/lib/mongodb/data lzuccarelli/mongodb:latest

  # start the microservices
  docker run -d --rm --link mongodb-service --name pubcode-golang -p 9000:9000 lzuccarelli/pubcode-golang:1.11.0 ./microservice
  docker run -d --rm --link pubcode-golang lzuccarelli/cronsftp-golang:1.11.0 ./microservice

fi
