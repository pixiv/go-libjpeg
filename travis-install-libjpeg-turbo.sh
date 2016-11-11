#!/bin/sh -eux

if [ "$1" != "" ]; then
  cd /tmp
  wget http://downloads.sourceforge.net/project/libjpeg-turbo/$1/libjpeg-turbo-official_$1_amd64.deb
  dpkg -x libjpeg-turbo-official_$1_amd64.deb /tmp/libjpeg-turbo-official
  mv /tmp/libjpeg-turbo-official/opt/libjpeg-turbo /tmp/libjpeg-turbo
fi
