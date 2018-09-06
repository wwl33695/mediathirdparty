//
// Created by Leo on 2016/10/9.
//

#include <stdio.h>
#include <string>
#include <stdlib.h>
#include <signal.h>
#include <unistd.h>
#include <string.h>
#include <sys/time.h>

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
std::string codec = "omxh264enc";

/*
 * RGB, BGR, RGBx, xRGB, BGRx, xBGR, RGBA, ARGB, BGRA, ABGR, I420, NV12, NV21, YV12, YUY2 can be supported on gst010
 * RGB, BGR, RGBx, xRGB, BGRx, xBGR, RGBA, ARGB, BGRA, ABGR, I420 can be supported on x264
 * RGB, BGR, RGBx, xRGB, BGRx, xBGR, RGBA, ARGB, BGRA, ABGR will product high cost on color space convert
 */
std::string pixfmt = "I420";
int width = 1280;
int height = 800;
int fps = 25;
int bitrate = 1024000; //unit: byte per second(B/s)
int iframeinterval = 25;
std::string host = "127.0.0.1";
int port = 9001;

size_t size = (size_t)(width * height * 1.5);

IEncoder *encoder;
int seq = 0;

/*
 * You need push your raw data in this function, the raw data format must be
 * same as your declaration in GSTEncoder Start(...) function.
 *
 * You should not apply for new memory and please set the data directly
 */
void OnNeedData(unsigned char *data, int64_t timestamp) {
    printf("%llu\n", (unsigned long long) timestamp);

    //This is a test frame, user can use memcpy(...) to push an image frame
    seq = (seq + 1) % 256;
    // Y
    for (int y = 0; y < height; y++) {
        for (int x = 0; x < width; x++) {
            data[y * width + x] = (unsigned char) (x + y + seq * 3);
        }
    }

    // U and V
    for (int y = 0; y < height / 2; y++) {
        for (int x = 0; x < width / 2; x++) {
            data[width * height + y * width / 2 + x] = (unsigned char) (128 + y + seq * 2);
            data[width * height + width * height / 4 + y * width / 2 + x] = (unsigned char) (64 + x + seq * 5);
        }
    }
}

void RuntimeErrorFunc(const char *message)
{
    printf("[ERROR] %s\n", message);
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

pthread_t encoder_thread_;

void *encoder_thread(void *arg)
{
    encoder->Start(codec, pixfmt, width, height, size, fps, bitrate, iframeinterval, host, port, OnNeedData, RuntimeErrorFunc);
    return NULL;
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

    pthread_create(&encoder_thread_, NULL, encoder_thread, NULL);
    pthread_join(encoder_thread_, NULL);

    delete encoder;
    encoder = NULL;
}
