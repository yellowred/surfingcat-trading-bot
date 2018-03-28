#!/bin/sh

if [[ -z "${SFTB_ROOT}" ]]; then
  echo "SFTB_ROOT is not set. Using ${HOME}/Sites/."
  SFTB_ROOT="${HOME}/Sites/go_hub/src/github.com/yellowred/surfingcat-trading-bot"
else
  echo "SFTB_ROOT=${SFTB_ROOT}."
fi

docker_install() {
  if [ ! -d "/Applications/Docker.app" ]; then
    echo "Downloading Docker ..."
    wget -q https://download.docker.com/mac/edge/Docker.dmg -O ~/Downloads/Docker.dmg
    hdiutil detach /Volumes/Docker
    hdiutil attach ~/Downloads/Docker.dmg
    echo "Copying Docker ..."
    cp -R /Volumes/Docker/Docker.app /Applications/Docker.app
    hdiutil detach /Volumes/Docker
    open /Applications/Docker.app
    echo "Done. Please enable Kubernetes in Docker for Mac. Then proceed to the cluster install."
  else
    echo "Docker for Mac is already installed."
  fi
}

sftb_workspace() {
    helm upgrade --install sftb ${SFTB_ROOT}/cd-assets/surfingcat-trading-bot/ --set trading-spa.persistence.enabled=true
}

purge() {
    helm delete --purge bgl
}

trading_server_reload() {
    docker build -t trading-server ${SFTB_ROOT}/server
    kubectl scale --replicas=0 deployment sftb-trading-server
    sleep 1
    kubectl scale --replicas=1 deployment sftb-trading-server
}


case $1 in
  docker) docker_install ;;
  sftb_workspace) sftb_workspace ;;
  purge) purge ;;
  trading_server_reload) trading_server_reload ;;
  *) echo "Usage: ./operations.sh <command>
Available commands: 
 * docker - download and install Docker for Mac,
 * purge - remove installed charts,
 * sftb_workspace - install sftb,
 * trading_server_reload - rebuild and reload trading-server";;
esac