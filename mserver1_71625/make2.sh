clear

set OLDGOPATH = $GOPATH
set OLDCGO_ENABLED=0 
set OLDGOOS=linux 
set OLDGOARCH=amd64

export GOPATH=`pwd`
#go get .
export CGO_ENABLED=0 
export GOOS=linux 
export GOARCH=amd64

#go test github.com/deepglint/dgmf/mserver/protocols/gb28181

#go env 
go build

export GOPATH=$OLDGOPATH
export CGO_ENABLED=$OLDCGO_ENABLED 
export GOOS=$OLDGOOS
export GOARCH=$OLDGOARCH

unset OLDGOPATH
unset OLDCGO_ENABLED
unset OLDGOOS
unset OLDGOARCH
