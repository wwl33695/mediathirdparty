//
// Created by Leo on 2016/10/9.
//


#ifndef MENC_GST010ENCODER_H
#define MENC_GST010ENCODER_H

#include "IEncoder.h"

#include <gst/gst.h>
#include <netinet/in.h>

class GST010Encoder: public IEncoder{
public:
    GST010Encoder();
    virtual ~GST010Encoder();

    void Start(std::string codec, std::string pixfmt, int width, int height, size_t size, int fps, int bitrate, int iframeinterval, std::string host, int port, NeedData inputcb, RuntimeError runtimeerr);
    void Stop();

private:
    static void OnOutDataInter(GstElement *sink, GST010Encoder *this_);
    static gboolean OnBusCallInter(GstBus * bus, GstMessage * message, gpointer data);

    GstElement *pipeline_;
    GstBus *bus_;
    struct sockaddr_in addr_;
    int sockfd_;
    int addr_len_;
    std::string codec_;
    bool running_;
};


#endif //MENC_GST010ENCODER_H