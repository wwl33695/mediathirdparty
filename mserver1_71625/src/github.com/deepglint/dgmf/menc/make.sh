#!/bin/bash

if [ "$1" != "x264" ] && [ "$1" != "gst010" ]
then
	echo "Usage: ./make.sh [x264 | gst010]"
fi

if [ "$1" = "x264" ]
then
	make x264
	mkdir -p MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)
	mkdir -p MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/lib
	mkdir -p MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/include

	cp build/lib/libmenc-x264.a MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/lib/
	cp README.md MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/
	cp ../VERSION MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/
	cp src/X264Encoder.h MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/include/
	cp src/IEncoder.h MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/include/
	tar -czvf MEnc-$(uname)-$(uname -m)-$(cat ../VERSION).tar.gz MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/
	rm -rf run.sh
	echo "./build/bin/menc-x264" >> run.sh
	chmod +x run.sh
fi

if [ "$1" = "gst010" ]
then
	make gst010
	mkdir -p MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)
	mkdir -p MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/lib
	mkdir -p MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/include

	cp build/lib/libmenc-gst010.a MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/lib/
	cp README.md MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/
	cp ../VERSION MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/
	cp src/GST010Encoder.h MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/include/
	cp src/IEncoder.h MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/include/
	tar -czvf MEnc-$(uname)-$(uname -m)-$(cat ../VERSION).tar.gz MEnc-$(uname)-$(uname -m)-$(cat ../VERSION)/
	
	rm -rf run.sh
	echo "./build/bin/menc-gst010" >> run.sh
	chmod +x run.sh
fi
