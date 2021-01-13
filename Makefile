GITTAG := $(shell git describe --tags --abbrev=0 2> /dev/null)
VERSION := $(or ${GITTAG}, v0.0.0)

all:
	rm -rf /opt/hvs-flavortemplates
	cp -r flavortemplates /opt/hvs-flavortemplates
	go build -ldflags "-X main.BuildVersion=$(VERSION)" -o flavor-convert-$(VERSION)
