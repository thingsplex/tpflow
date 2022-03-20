version="1.3.8"
version_file=VERSION
working_dir=$(shell pwd)
arch="armhf"
remote_host = "fh@futurehome-smarthub.local"
reprepo_host = "reprepro@archive.futurehome.no"

clean:
	-rm tpflow

build-go:
	go build -o tpflow cmd/main.go

build-go-arm:
	GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o tpflow cmd/main.go

build-go-amd:
	GOOS=linux GOARCH=amd64 go build -o tpflow cmd/main.go


configure-arm:
	python ./scripts/config_env.py prod $(version) armhf

configure-amd64:
	python ./scripts/config_env.py prod $(version) amd64


package-tar:
	tar cvzf tpflow_$(version).tar.gz tpflow VERSION

package-deb-doc:
	@echo "Packaging application as debian package"
	find package/ -name ".DS_Store" -type f -delete
	chmod a+x package/debian/DEBIAN/*
	cp tpflow package/debian/opt/thingsplex/tpflow/
	cp -R extlibs package/debian/opt/thingsplex/tpflow/
	cp VERSION package/debian/opt/thingsplex/tpflow/
	mkdir -p package/debian/opt/thingsplex/tpflow/var/connectors/data/httpserver/static
	cp static package/debian/opt/thingsplex/tpflow/var/connectors/data/httpserver/
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian
	@echo "Done"


tar-arm: build-js build-go-arm package-deb-doc
	@echo "The application was packaged into tar archive "

deb-arm : clean configure-arm build-go-arm package-deb-docHemmelig1
	mv package/debian.deb package/build/tpflow_$(version)_armhf.deb

deb-amd : configure-amd64 build-go-amd package-deb-doc
	mv debian.deb tpflow_$(version)_amd64.deb

upload :
	scp package/build/tpflow_$(version)_armhf.deb $(remote_host):~/

upload-install : upload
	ssh -t $(remote_host) "sudo dpkg -i tpflow_$(version)_armhf.deb"

remote-install : deb-arm upload
	ssh -t $(remote_host) "sudo dpkg -i tpflow_$(version)_armhf.deb"

run :
	go run cmd/main.go -c testdata/var/config.json

publish-reprepo:
	scp package/build/tpflow_$(version)_armhf.deb $(reprepo_host):~/apps

# Docker
build-go-amd64:
	cd GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${version}" -mod=vendor -o package/docker/build/amd64/tpflow cmd/main.go

build-go-arm64:
	cd GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${version}" -mod=vendor -o package/docker/build/arm64/tpflow cmd/main.go

package-docker-local: build-go-amd64
	docker build --build-arg TARGETARCH=amd64 -t thingsplex/tpflow:${version} -t thingsplex/tpflow:latest ./package/docker

package-docker-multi:
	docker buildx build --platform linux/arm64,linux/amd64 -t thingsplex/tpflow:${version} -t thingsplex/tpflow:latest ./package/docker --push

docker-multi-setup : configure-amd64
	docker buildx create --name mybuilder
	docker buildx use mybuilder

docker-multi-publish : build-go-arm64 build-go-amd64 package-docker-multi

run-docker:
	docker run -d -v tpflow:/thingsplex/tpflow/var --restart unless-stopped \
    -e MQTT_URI=tcp://192.168.86.33:1884 -e MQTT_USERNAME=uname -e MQTT_PASSWORD=pass \
    --network tplex-net --name tpflow thingsplex/tpflow:latest

init-mqtt:
	 docker run -d --name mqtt -p 1883:1883 -p 9001:9001 -v $(working_dir)/testdata/mosquitto/config:/mosquitto/config eclipse-mosquitto:2.0.14

start-mqtt:
	docker start mqtt

.phony : clean
