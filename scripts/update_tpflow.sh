#!/bin/bash
version=$1
echo "Updating to version: $version"
wget https://storage.googleapis.com/fh-repo/tpflow_${version}_armhf.deb
sudo dpkg -i tpflow_${version}_armhf.deb
