all:
	rm -rf /opt/hvs-flavortemplates
	cp -r flavortemplates /opt/hvs-flavortemplates
	go build