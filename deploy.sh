#!/bin/bash
set -e
msg ()
{
    msg=$1
    echo ""
    echo ">>>>> $msg <<<<<"
    echo ""
}
msg "Heroku login"
heroku container:login
msg "Building ws image"
docker build -t "luax/ws" -f "services/Dockerfile" --build-arg SERVICE_NAME="ws" .

msg "Building gateway image"
docker build -t "luax/gateway" -f "services/Dockerfile" --build-arg SERVICE_NAME="gateway" .

msg "Deploying ws"
cd heroku
cd ws
heroku container:push --recursive web -a luax-test-ws
heroku container:release web -a luax-test-ws
cd ..
cd gateway
msg "Deploying gateway"
heroku container:push --recursive web -a luax-test-gateway
heroku container:release web -a luax-test-gateway
msg "Done"
