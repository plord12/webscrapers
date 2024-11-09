#!/usr/bin/with-contenv bashio

export HA_USERNAME=$(bashio::config 'username')
export HA_PASSWORD=$(bashio::config 'password')

mkdir -p ${HOME}
cd ${HOME}

/ha_ss-linux-arm64 -restport $(bashio::addon.port 3500)