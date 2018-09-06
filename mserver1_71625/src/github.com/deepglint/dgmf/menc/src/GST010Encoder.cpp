//
// Created by Leo on 2016/10/9.
//

#include "GST010Encoder.h"

#include <stdlib.h>
#include <string>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <string.h>
#include <gst/gst.h>
#include <unistd.h>
#include <sys/time.h>
#include <stdio.h>
#include <gst/app/gstappsrc.h>

#include "gstappsink.h"

GST010Encoder::GST010Encoder() {
    gst_init(NULL, NULL);
    pipeline_ = NULL;
    sockfd_ = 0;
    addr_len_ = 0;
    running_ = false;
}

GST010Encoder::~GST010Encoder() {
    this->Stop();
}

void GST010Encoder::OnOutDataInter(GstElement *sink, GST010Encoder *this_) {
    GstBuffer * buffer = gst_app_sink_pull_buffer(GST_APP_SINK(sink));
    unsigned char *data = GST_BUFFER_DATA(buffer);
    int size = GST_BUFFER_SIZE(buffer);

    if (this_->codec_.compare("x264enc") == 0) {

        if (size>=5 && data[4] == 0x09 && data[5] == 0x30) {
            data+=5;
            size-=5;
            data[0] = 0x00;
        } else if (size>=5 && data[4] == 0x09 && data[5] == 0x10) {
            data+=5;
            for (int i = 0; i < size-5-4; i++){
                if (data[i+1] == 0x00 && data[i+2] == 0x00 && data[i+3] == 0x01 && data[i+4] == 0x65){
                    data[i] = 0x00;
                    break;
                } else{
                    data[i] = data[i+1];
                }
            }
        }
    }
    for (int i = 0; i <= size / 1440; i++) {
        if (i == size / 1440) {
            if (size%1440 != 0) {
                sendto(this_->sockfd_, data + i * 1440, size % 1440, 0,
                       (struct sockaddr *) &this_->addr_, this_->addr_len_);
            }
        } else {
            sendto(this_->sockfd_, data+i*1440, 1440, 0, (struct sockaddr *)&this_->addr_, this_->addr_len_);
        }
    }
    gst_buffer_unref(buffer);
}

gboolean GST010Encoder::OnBusCallInter(GstBus * bus, GstMessage * message, gpointer data)
{
    RuntimeError runtimeerr = (RuntimeError)data;
    if (GST_MESSAGE_TYPE(message) == GST_MESSAGE_ERROR) {
        runtimeerr(gst_message_type_get_name(GST_MESSAGE_TYPE(message)));
    }
    return true;
}

void GST010Encoder::Start(std::string codec, std::string pixfmt, int width, int height, size_t size, int fps, int bitrate,
                       int iframeinterval, std::string host, int port, NeedData inputcb, RuntimeError runtimeerr) {

        addr_len_ = sizeof(struct sockaddr_in);

        sockfd_ = socket(AF_INET, SOCK_DGRAM, 0);
        addr_.sin_family = AF_INET;
        addr_.sin_port = htons(port);
        addr_.sin_addr.s_addr = inet_addr(host.c_str());
        codec_ = codec;

        std::string ps = "appsrc name=leosrc";
        if (pixfmt.compare("RGB") == 0) {
            ps += " ! video/x-raw-rgb, bpp=24, depth=24, endianness=4321, red_mask=16711680, green_mask=65280, blue_mask=255";
        } else if (pixfmt.compare("BGR") == 0) {
            ps += " ! video/x-raw-rgb, bpp=24, depth=24, endianness=4321, red_mask=255, green_mask=65280, blue_mask=16711680";
        } else if (pixfmt.compare("RGBx") == 0) {
            ps += " ! video/x-raw-rgb, bpp=32, depth=24, endianness=4321, red_mask=-16777216, green_mask=16711680, blue_mask=65280";
        } else if (pixfmt.compare("xRGB") == 0) {
            ps += " ! video/x-raw-rgb, bpp=32, depth=24, endianness=4321, red_mask=16711680, green_mask=65280, blue_mask=255";
        } else if (pixfmt.compare("BGRx") == 0) {
            ps += " ! video/x-raw-rgb, bpp=32, depth=24, endianness=4321, red_mask=65280, green_mask=16711680, blue_mask=-16777216";
        } else if (pixfmt.compare("xBGR") == 0) {
            ps += " ! video/x-raw-rgb, bpp=32, depth=24, endianness=4321, red_mask=255, green_mask=65280, blue_mask=16711680";
        } else if (pixfmt.compare("RGBA") == 0) {
            ps += " ! video/x-raw-rgb, bpp=32, depth=32, endianness=4321, red_mask=-16777216, green_mask=16711680, blue_mask=65280, alpha_mask=255";
        } else if (pixfmt.compare("ARGB") == 0) {
            ps += " ! video/x-raw-rgb, bpp=32, depth=32, endianness=4321, red_mask=16711680, green_mask=65280, blue_mask=255, alpha_mask=-16777216";
        } else if (pixfmt.compare("BGRA") == 0) {
            ps += " ! video/x-raw-rgb, bpp=32, depth=32, endianness=4321, red_mask=65280, green_mask=16711680, blue_mask=-16777216, alpha_mask=255";
        } else if (pixfmt.compare("ABGR") == 0) {
            ps += " ! video/x-raw-rgb, bpp=32, depth=32, endianness=4321, red_mask=255, green_mask=65280, blue_mask=16711680, alpha_mask=-16777216";
        } else if (pixfmt.compare("I420") == 0) {
            ps += " ! video/x-raw-yuv, format=(fourcc)I420";
        } else if (pixfmt.compare("NV12") == 0) {
            ps += " ! video/x-raw-yuv, format=(fourcc)NV12";
        } else if (pixfmt.compare("NV21") == 0) {
            ps += " ! video/x-raw-yuv, format=(fourcc)NV21";
        } else if (pixfmt.compare("YV12") == 0) {
            ps += " ! video/x-raw-yuv, format=(fourcc)YV12";
        } else if (pixfmt.compare("YUY2") == 0) {
            ps += " ! video/x-raw-yuv, format=(fourcc)YUY2";
        }
        ps += ", width=" + std::to_string(width) + ", height=" + std::to_string(height) + ", framerate=" +
              std::to_string(fps) + "/1";
        ps += " ! ffmpegcolorspace ! video/x-raw-yuv, format=(fourcc)I420, width=" + std::to_string(width) +
              ",height=" +
              std::to_string(height) + ",framerate=" + std::to_string(fps) + "/1 ! ";

        if (codec.compare("x264enc") == 0) {
            ps += "x264enc";
            ps += " tune=zerolatency speed-preset=ultrafast";
            ps += " bitrate=" + std::to_string(bitrate / 1000);
            ps += " key-int-max=" + std::to_string(iframeinterval);
            ps += " byte-stream=true";
        }

        if (codec.compare("omxh264enc") == 0) {
            ps += "nv_omx_h264enc low-latency=1";
            ps += " bitrate=" + std::to_string(bitrate);
            ps += " iframeinterval=" + std::to_string(iframeinterval);
            ps += " rc-mode=0";
            ps += " ! video/x-h264, stream-format=byte-stream";
            ps += ", width=" + std::to_string(width) + ", height=" + std::to_string(height) + ", framerate=" +
                  std::to_string(fps) + "/1";
        }

        ps += " ! appsink name=leosink";

        pipeline_ = gst_parse_launch(ps.c_str(), NULL);
        GstElement *appsrc = gst_bin_get_by_name(GST_BIN(pipeline_), "leosrc");
        GstElement *appsink = gst_bin_get_by_name(GST_BIN(pipeline_), "leosink");
        g_object_set(G_OBJECT(appsrc), "is-live", true, "format", GST_FORMAT_TIME, NULL);
        gst_app_src_set_stream_type(GST_APP_SRC(appsrc), GST_APP_STREAM_TYPE_STREAM);
        gst_app_src_set_max_bytes(GST_APP_SRC(appsrc), size*4);
        g_object_set(appsrc, "block", false, NULL);

        g_object_set(G_OBJECT(appsink), "emit-signals", true, NULL);
        g_signal_connect(G_OBJECT(appsink), "new-buffer", G_CALLBACK(OnOutDataInter), this);

        bus_ = gst_pipeline_get_bus(GST_PIPELINE(pipeline_));
        gst_bus_add_watch(bus_, OnBusCallInter, (void *)runtimeerr);
        gst_object_unref(bus_);

        gst_element_set_state(pipeline_, GST_STATE_PLAYING);

        running_ = true;
        GTimer *timer = g_timer_new();
        g_timer_start(timer);
        GstClockTime timestamp = 0;
        GstBuffer *buffer = gst_buffer_new_and_alloc(size);
        GstFlowReturn ret;

        while (running_) {
            struct timeval t0;
            gettimeofday(&t0, NULL);

            inputcb(GST_BUFFER_DATA(buffer), (int64_t) timestamp);
            GST_BUFFER_TIMESTAMP(buffer) = timestamp;
            GST_BUFFER_DURATION (buffer) = (unsigned int)(1000000000 / fps);
            timestamp += GST_BUFFER_DURATION(buffer);
            g_signal_emit_by_name(appsrc, "push-buffer", buffer, &ret);

            struct timeval t1;
            gettimeofday(&t1, NULL);

            int delta = (t1.tv_sec - t0.tv_sec) * 1000000 + (t1.tv_usec - t0.tv_usec);
            if ( delta > 0 && delta < 1000000/fps) {
                usleep(1000000/fps - delta);
            }
        }

        gst_buffer_unref(buffer);
        gst_element_set_state(pipeline_, GST_STATE_NULL);
        gst_object_unref(appsrc);
        gst_object_unref(pipeline_);
        sockfd_ = 0;
        addr_len_ = 0;
        codec_ = "";
}

void GST010Encoder::Stop() {
    running_= false;
}