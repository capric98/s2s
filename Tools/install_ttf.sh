#!/bin/bash
mkdir -p /usr/share/fonts/zh_CN/
mv msyh.ttf /usr/share/fonts/zh_CN/
chmod -R 766 /usr/share/fonts/zh_CN

apt-get update >> /dev/null
apt-get install -y xfonts-utils >> /dev/null
mkfontscale
mkfontdir
fc-cache -fv
reboot