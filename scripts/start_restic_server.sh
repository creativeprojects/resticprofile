#!/bin/sh

image=restic/rest-server
container=rest_server
data=/datapool/restic
options=""

docker pull ${image}
docker run -d -p 8000:8000 -v ${data}:/data --name ${container} --restart always -e "OPTIONS=${options}" ${image}
