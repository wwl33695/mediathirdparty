mkdir -p MServer-Darwin-x86_64-$(cat ../VERSION)
mkdir -p MServer-Linux-x86_64-$(cat ../VERSION)
mkdir -p MServer-Linux-armv7l-$(cat ../VERSION)

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o MServer-Darwin-x86_64-$(cat ../VERSION)/mserver
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o MServer-Linux-x86_64-$(cat ../VERSION)/mserver
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o MServer-Linux-armv7l-$(cat ../VERSION)/mserver

cp README.md MServer-Darwin-x86_64-$(cat ../VERSION)/
cp README.md MServer-Linux-x86_64-$(cat ../VERSION)/
cp README.md MServer-Linux-armv7l-$(cat ../VERSION)/

cp spout/MServer.postman_collection.json MServer-Darwin-x86_64-$(cat ../VERSION)/
cp spout/MServer.postman_collection.json MServer-Linux-x86_64-$(cat ../VERSION)/
cp spout/MServer.postman_collection.json MServer-Linux-armv7l-$(cat ../VERSION)/


cp ../VERSION MServer-Darwin-x86_64-$(cat ../VERSION)/
cp ../VERSION MServer-Linux-x86_64-$(cat ../VERSION)/
cp ../VERSION MServer-Linux-armv7l-$(cat ../VERSION)/

tar -czvf MServer-Darwin-x86_64-$(cat ../VERSION).tar.gz MServer-Darwin-x86_64-$(cat ../VERSION)/
tar -czvf MServer-Linux-x86_64-$(cat ../VERSION).tar.gz MServer-Linux-x86_64-$(cat ../VERSION)/
tar -czvf MServer-Linux-armv7l-$(cat ../VERSION).tar.gz MServer-Linux-armv7l-$(cat ../VERSION)/
