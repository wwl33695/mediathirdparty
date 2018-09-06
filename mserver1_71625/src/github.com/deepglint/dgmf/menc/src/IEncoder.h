//
// Created by Leo on 2016/11/23.
//

#ifndef MENC_IENCODER_H
#define MENC_IENCODER_H

#include <stdint.h>
#include <iostream>

typedef void (*NeedData)(unsigned char *data, int64_t timestamp);
typedef void (*RuntimeError)(const char* message);

class IEncoder {
public:
    IEncoder();
    virtual ~IEncoder();

    virtual void Start(std::string codec, std::string pixfmt, int width, int height, size_t size, int fps, int bitrate, int iframeinterval, std::string host, int port, NeedData inputcb, RuntimeError runtimeerr) = 0;
    virtual void Stop() = 0;

};


#endif //MENC_IENCODER_H
