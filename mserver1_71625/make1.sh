clear

set OLDGOPATH = $GOPATH

export GOPATH=`pwd`
#go get .

#go test github.com/deepglint/dgmf/mserver/protocols/gb28181

#go env 
go build

export GOPATH=$OLDGOPATH

unset OLDGOPATH