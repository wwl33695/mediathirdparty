# MEnc
## Overview
MEnc is a project to encode raw data(RGB or YUV) to h264 rtp stream. It's a part of DGMF. This project can be carried out on a stable encoding and data package.
## Dependency Libraries
You can choose make gst010 or x264 and the different target has different libraries.
### GST010 on Linux-x86_64 and Linux-armv7l
```
gstreamer0.10
gstreamer0.10 ffmpeg plugin
gstreamer0.10 base plugin
glib2.0
make (compile environment)
```
### GST010 on Darwin-x86_64
```
gstreamer0.10
gstreamer0.10 ffmpeg plugin
gstreamer0.10 base plugin
gstreamer0.10 ugly plugin
glib2.0
make (compile environment)
```
### X264 on Linux-x86_64, Linux-armv7l and Darwin-x86_64
```
x264
make (compile environment)
```

### How to install Gstreamer0.10 on Ubuntu x86_64
```
$ add-apt-repository -y ppa:mc3man/gstffmpeg-keep
$ apt-get update
$ apt-get -y --force-yes install gstreamer0.10
$ apt-get -y --force-yes install gstreamer0.10-ffmpeg
$ apt-get -y --force-yes install libgstreamer-plugins-base0.10-dev
```
### How to install Gstreamer0.10 on Ubuntu armv7l
```
$ apt-get update
$ apt-get -y --force-yes install gstreamer0.10
$ apt-get -y --force-yes install libgstreamer-plugins-base0.10-dev
$ apt-get -y --force-yes install libgstreamer-plugins-ugly0.10-dev
```
### How to install Gstreamer0.10 on MacOSX x86_64
```
$ brew install gstreamer010
$ brew install gst-plugins-base010
$ brew install gst-plugins-ugly010 --with-x264
$ brew install gst-ffmpeg010
```
### How to install x264 on Ubuntu x86_64 and Ubuntu armv7l 
```
$ apt-get update
$ apt-get -y --force-yes install x264
```
### How to install x264 on MacOSX x86_64
```
$ brew install x264
```
## How to build
```
$ cd menc
$ ./make
```
You can find quick start demo at menc/build/bin/menc, and static library at menc/build/lib/libmenc.a
## Quick Start

```
//
// Created by Leo on 2016/10/9.
//

#include <stdio.h>
#include <string>
#include <stdlib.h>
#include <signal.h>

#ifdef USE_GST010
#include "GST010Encoder.h"
#endif
#ifdef USE_X264
#include "X264Encoder.h"
#endif
#include "IEncoder.h"

/*
 * x264enc for software encoding on Darwin-x86_64, Linux-x86_64 and Linux-armv7l
 * omxh264enc for hardware encoding on Linux-armv7l only
 */
std::string codec = "x264enc";

/*
 * RGB, BGR, RGBx, xRGB, BGRx, xBGR, RGBA, ARGB, BGRA, ABGR, I420, NV12, NV21, YV12, YUY2 can be supported
 * RGB, BGR, RGBx, xRGB, BGRx, xBGR, RGBA, ARGB, BGRA, ABGR will product high cost on color space convert
 */
std::string pixfmt = "I420";
int width = 640;
int height = 480;
int fps = 25;
int bitrate = 1024000; //unit: byte per second(B/s)
int iframeinterval = 25;
std::string host = "127.0.0.1";
int port = 9004;

size_t size = (size_t)(width * height * 3 / 2);

IEncoder *encoder;
int seq = 0;

/*
 * You need push your raw data in this function, the raw data format must be
 * same as your declaration in GSTEncoder Start(...) function.
 *
 * You should not apply for new memory and please set the data directly
 */
void OnNeedData(unsigned char *data, int64_t timestamp) {
    printf("%llu\n", (unsigned long long)timestamp);

    //This is a test frame, user can use memcpy(...) to push an image frame
    seq = (seq + 1) % 256;
    // Y
    for (int y = 0; y < height; y++) {
        for (int x = 0; x < width; x++) {
            data[y * width + x] = (unsigned char) (x + y + seq * 3);
        }
    }

    // U and V
    for (int y = 0; y < height/2; y++) {
        for (int x = 0; x < width/2; x++) {
            data[width*height + y * width/2 + x] = (unsigned char) (128 + y + seq * 2);
            data[width*height+width*height/4 + y * width/2 + x] = (unsigned char) (64 + x + seq * 5);
        }
    }
}

/*
 * You can receive a system signal. In this demo, I will show you how to shutdown an encoder
 * and delete memory safely.
 *
 * Please do not free memory after stop function, because, stop is asynchronous, you can free it
 * after start function.
 */
void OnSignalReceived(int signum) {
    // You can stop encoder, change resolution or fps ... and restart the encoder
    encoder->Stop();

    // This is wrong!!! Don't do this!!!
    // delete encoder;
}

int main() {

    // Register a call back function to receive system signal
    signal(SIGINT, OnSignalReceived);

#ifdef USE_GST010
    encoder = new GST010Encoder();
#endif
#ifdef USE_X264
    encoder = new X264Encoder();
#endif
    encoder->Start(codec, pixfmt, width, height, size, fps, bitrate, iframeinterval, host, port, OnNeedData);
    delete encoder;
    encoder = NULL;
}

```
