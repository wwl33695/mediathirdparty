//
// Created by Leo on 2016/11/23.
//

#ifndef MENC_X264ENCODER_H
#define MENC_X264ENCODER_H

#include "IEncoder.h"

#include <netinet/in.h>
#include <x264.h>
#include <string>

class X264Encoder : public IEncoder{
public:
    X264Encoder();
    virtual ~X264Encoder();

    void Start(std::string codec, std::string pixfmt, int width, int height, size_t size, int fps, int bitrate, int iframeinterval, std::string host, int port, NeedData inputcb, RuntimeError runtimeerr);
    void Stop();

private:
    void EncodeHandler();

    bool running_;
    unsigned char *raw_data_;
    unsigned char *yuv_data_;
    unsigned char *h264_data_;
    int64_t timestamp_;
    NeedData inputcb_;
    int fps_;
    std::string pixfmt_;
    x264_picture_t *pic_in_;
    x264_picture_t *pic_out_;
    int width_;
    int height_;
    x264_t *x264_handle_;
    int sockfd_;
    struct sockaddr_in addr_;
    int addr_len_;
};


#endif //MENC_X264ENCODER_H