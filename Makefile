# The import path is where your repository can be found.
# To import subpackages, always prepend the full import path.
# If you change this, run `make clean`. Read more: https://git.io/vM7zV
IMPORT_PATH := loadbalance

V := 1 # When V is set, print commands and build progress.

export GOPATH := $(CURDIR)/.GOPATH
export CGO_ENABLED := 0
unexport GOBIN

# Space separated patterns of packages to skip in list, test, format.
IGNORED_PACKAGES := /vendor/

all: agent

agent: .GOPATH/.ok
	$Q go install -tags netgo $(IMPORT_PATH)

update: .GOPATH/.ok
	$Q glide up -v
	
clean:
	$Q rm -rf bin pkg .GOPATH

Q := $(if $V,,@)

.GOPATH/.ok:
	$Q rm -rf $(GOPATH)
	$Q mkdir -p $(GOPATH)/src
	$Q ln -sf $(CURDIR) $(GOPATH)/src/$(IMPORT_PATH)
	$Q mkdir -p $(CURDIR)/pkg
	$Q ln -sf $(CURDIR)/pkg $(GOPATH)/pkg
	$Q mkdir -p $(CURDIR)/bin
	$Q ln -sf $(CURDIR)/bin $(GOPATH)/bin
	$Q touch $@