version="1.2.2"
version_file=VERSION
working_dir=$(shell pwd)
arch="armhf"
remote_host = "fh@cube.local"
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
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian
	@echo "Done"


tar-arm: build-js build-go-arm package-deb-doc
	@echo "The application was packaged into tar archive "

deb-arm : clean configure-arm build-go-arm package-deb-doc
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

.phony : clean
