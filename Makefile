NAME = argeos
FILES_TO_RPM = argeos systemd/argeos.service
SPECFILE = argeos.spec
PACKAGE  = $(shell awk '$$1 == "Name:"     { print $$2 }' $(SPECFILE) )
VERSION  = $(shell awk '$$1 == "Version:"  { print $$2 }' $(SPECFILE) )
RELEASE  = $(shell awk '$$1 == "Release:"  { print $$2 }' $(SPECFILE) )
OS_ARCH = $(shell echo "$$(uname -s)/$$(uname -m)")
DIST               ?= $(shell rpm --eval %{dist})
rpmbuild = ${shell pwd}/build
MAIN = ./cmd/.

.PHONY: build

default: build

build:
	@go build -o $(NAME) $(MAIN)

debug:
	@go build -gcflags=all="-N -l" -o $(NAME) $(MAIN)

clean:
	@rm -rf $(PACKAGE)-$(VERSION)
	@rm -rf $(rpmbuild)
	@rm -rf *rpm
	@rm -rf $(NAME)

rpmdefines=--define='_topdir ${rpmbuild}' \
        --define='_sourcedir %{_topdir}/SOURCES' \
        --define='_builddir %{_topdir}/BUILD' \
        --define='_srcrpmdir %{_topdir}/SRPMS' \
        --define='_rpmdir %{_topdir}/RPMS' \
		--define='dist $(DIST)'

dist: clean build
	@mkdir -p $(PACKAGE)-$(VERSION)
	@cp -r $(FILES_TO_RPM) $(PACKAGE)-$(VERSION)
	tar cpfz ./$(PACKAGE)-$(VERSION).tar.gz $(PACKAGE)-$(VERSION)

prepare: dist
	@mkdir -p $(rpmbuild)/RPMS/x86_64
	@mkdir -p $(rpmbuild)/SRPMS/
	@mkdir -p $(rpmbuild)/SPECS/
	@mkdir -p $(rpmbuild)/SOURCES/
	@mkdir -p $(rpmbuild)/BUILD/
	@mv $(PACKAGE)-$(VERSION).tar.gz $(rpmbuild)/SOURCES 
	@cp $(SPECFILE) $(rpmbuild)/SOURCES 

srpm: prepare $(SPECFILE)
	rpmbuild --nodeps -bs $(rpmdefines) $(SPECFILE)

rpm: srpm
	rpmbuild --nodeps -bb $(rpmdefines) $(SPECFILE)
	cp $(rpmbuild)/RPMS/x86_64/* .
