#!/usr/bin/with-contenv bashio

export HA_USERNAME=$(bashio::config 'username')
export HA_PASSWORD=$(bashio::config 'password')
export HEADLESS=true

mkdir -p ${HOME}
cd ${HOME}

/ha_ss --restport $(bashio::addon.port 3500)