//
// Created by Leo on 2016/11/23.
//

#include "X264Encoder.h"

#include <stdlib.h>
#include <time.h>
#include <signal.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <stdio.h>
#include <string.h>
#include <cv.h>

X264Encoder::X264Encoder() {
    sockfd_ = 0;
    addr_len_ = 0;
    running_ = false;

    inputcb_ = NULL;
    raw_data_ = NULL;
    yuv_data_ = NULL;
    h264_data_ = NULL;
    timestamp_ = 0;
    fps_ = 0;
    pic_in_ = NULL;
    pic_out_ = NULL;
    x264_handle_ = NULL;
    pixfmt_ = "";
}

X264Encoder::~X264Encoder() {
    this->Stop();
}

void RGB2I420(unsigned char *raw_data, unsigned char *yuv_data, int width, int height) {
    cv::Mat src(height, width, CV_8UC3, raw_data);
    cv::Mat dst(height + height / 2, width, CV_8UC1, yuv_data);
    cv::cvtColor(src, dst, CV_RGB2YUV_I420);
}

void BGR2I420(unsigned char *raw_data, unsigned char *yuv_data, int width, int height) {
    cv::Mat src(height, width, CV_8UC3, raw_data);
    cv::Mat dst(height + height / 2, width, CV_8UC1, yuv_data);
    cv::cvtColor(src, dst, CV_BGR2YUV_I420);
}

void RGBA2I420(unsigned char *raw_data, unsigned char *yuv_data, int width, int height) {
    cv::Mat src(height, width, CV_8UC4, raw_data);
    cv::Mat dst(height + height / 2, width, CV_8UC1, yuv_data);
    cv::cvtColor(src, dst, CV_RGBA2YUV_I420);
}

void BGRA2I420(unsigned char *raw_data, unsigned char *yuv_data, int width, int height) {
    cv::Mat src(height, width, CV_8UC4, raw_data);
    cv::Mat dst(height + height / 2, width, CV_8UC1, yuv_data);
    cv::cvtColor(src, dst, CV_BGRA2YUV_I420);
}

void X264Encoder::EncodeHandler() {
    x264_nal_t *nals = NULL;
    int inal = 0;

    this->inputcb_(this->raw_data_, (int64_t)(this->timestamp_ * 1000000000.0 / (double) this->fps_));
    this->timestamp_ += 1;
    this->pic_in_->i_pts = this->timestamp_;

    if (strcmp(this->pixfmt_.c_str(), "RGB") == 0) {
        RGB2I420(this->raw_data_, this->yuv_data_, this->width_, this->height_);
    } else if (strcmp(this->pixfmt_.c_str(), "BGR") == 0) {
        BGR2I420(this->raw_data_, this->yuv_data_, this->width_, this->height_);
    } else if (strcmp(this->pixfmt_.c_str(), "RGBA") == 0 || strcmp(this->pixfmt_.c_str(), "RGBx") == 0) {
        RGBA2I420(this->raw_data_, this->yuv_data_, this->width_, this->height_);
    } else if (strcmp(this->pixfmt_.c_str(), "BGRA") == 0 || strcmp(this->pixfmt_.c_str(), "BGRx") == 0) {
        BGRA2I420(this->raw_data_, this->yuv_data_, this->width_, this->height_);
    } else if (strcmp(this->pixfmt_.c_str(), "ARGB") == 0 || strcmp(this->pixfmt_.c_str(), "xRGB") == 0) {
        for (int i = 0; i < this->width_*4; i+=4) {
            for (int j = 0; j < this->height_; j++) {
                uint8_t tmp = 0;
                tmp = this->raw_data_[j*this->width_*4+i];
                this->raw_data_[j*this->width_*4+i] = this->raw_data_[j*this->width_*4+i+3];
                this->raw_data_[j*this->width_*4+i+3] = tmp;

                tmp = this->raw_data_[j*this->width_*4+i+1];
                this->raw_data_[j*this->width_*4+i+1] = this->raw_data_[j*this->width_*4+i+2];
                this->raw_data_[j*this->width_*4+i+2] = tmp;
            }
        }
        BGRA2I420(this->raw_data_, this->yuv_data_, this->width_, this->height_);
    } else if (strcmp(this->pixfmt_.c_str(), "ABGR") == 0 || strcmp(this->pixfmt_.c_str(), "xBGR") == 0){
        for (int i = 0; i < this->width_*4; i+=4) {
            for (int j = 0; j < this->height_; j++) {
                uint8_t tmp = 0;
                tmp = this->raw_data_[j*this->width_*4+i];
                this->raw_data_[j*this->width_*4+i] = this->raw_data_[j*this->width_*4+i+3];
                this->raw_data_[j*this->width_*4+i+3] = tmp;

                tmp = this->raw_data_[j*this->width_*4+i+1];
                this->raw_data_[j*this->width_*4+i+1] = this->raw_data_[j*this->width_*4+i+2];
                this->raw_data_[j*this->width_*4+i+2] = tmp;
            }
        }
        RGBA2I420(this->raw_data_, this->yuv_data_, this->width_, this->height_);
    } else {
        memcpy(this->yuv_data_, this->raw_data_, this->width_ * this->height_ * 3 / 2);
    }

    this->pic_in_->img.plane[0] = this->yuv_data_;
    this->pic_in_->img.plane[1] = this->yuv_data_ + this->width_ * this->height_;
    this->pic_in_->img.plane[2] =
            this->yuv_data_ + this->width_ * this->height_ + this->width_ * this->height_ / 4;

    x264_encoder_encode(this->x264_handle_, &nals, &inal, this->pic_in_, this->pic_out_);

    int size = 0;
    for (int i = 0; i < inal; ++i) {
        memcpy(this->h264_data_ + size, nals[i].p_payload, nals[i].i_payload);
        size += nals[i].i_payload;
    }

    for (int i = 0; i <= size / 1440; i++) {
        if (i == size / 1440) {
            if (size % 1440 != 0) {
                sendto(this->sockfd_, this->h264_data_ + i * 1440, size % 1440, 0,
                       (struct sockaddr *) &this->addr_, this->addr_len_);
            }
        } else {
            sendto(this->sockfd_, this->h264_data_ + i * 1440, 1440, 0, (struct sockaddr *) &this->addr_,
                   this->addr_len_);
        }
    }
}

void X264Encoder::Start(std::string codec, std::string pixfmt, int width, int height, size_t size, int fps, int bitrate,
                        int iframeinterval, std::string host, int port, NeedData inputcb, RuntimeError runtimeerr) {
    inputcb_ = inputcb;
    fps_ = fps;
    width_ = width;
    height_ = height;
    pixfmt_ = pixfmt;

    addr_len_ = sizeof(struct sockaddr_in);

    sockfd_ = socket(AF_INET, SOCK_DGRAM, 0);
    bzero(&addr_, sizeof(addr_));
    addr_.sin_family = AF_INET;
    addr_.sin_port = htons(port);
    addr_.sin_addr.s_addr = inet_addr(host.c_str());

    raw_data_ = (unsigned char *) malloc(size);
    h264_data_ = (unsigned char *) malloc(size);
    yuv_data_ = (unsigned char *) malloc(width * height * 3 / 2);
    timestamp_ = 0;

    pic_in_ = (x264_picture_t *) malloc(sizeof(x264_picture_t));
    pic_out_ = (x264_picture_t *) malloc(sizeof(x264_picture_t));
    x264_param_t *pParam = (x264_param_t *) malloc(sizeof(x264_param_t));

    x264_param_default_preset(pParam, "veryfast", "zerolatency");
    pParam->i_width = width;
    pParam->i_height = height;
    pParam->i_keyint_max = iframeinterval;
    pParam->i_bframe = 0;
    pParam->i_fps_den = 1;
    pParam->i_fps_num = fps;
    pParam->i_timebase_den = pParam->i_fps_num;
    pParam->i_timebase_num = pParam->i_fps_den;
    pParam->rc.i_bitrate = bitrate;
    pParam->rc.b_mb_tree = 0;
    pParam->i_csp = X264_CSP_I420;
    x264_param_apply_profile(pParam, "baseline");

    x264_handle_ = x264_encoder_open(pParam);
    x264_picture_init(pic_out_);
    x264_picture_alloc(pic_in_, X264_CSP_I420, pParam->i_width, pParam->i_height);

    struct timeval tv;
    tv.tv_sec = 0;
    tv.tv_usec = 1000;

    running_ = true;
    while (running_ == true) {
        struct timeval t0;
        struct timeval t1;
        gettimeofday(&t0, NULL);
        EncodeHandler();
        while (1) {
            gettimeofday(&t1, NULL);
            int delta = (t1.tv_sec * 1000 + t1.tv_usec / 1000) - (t0.tv_sec * 1000 + t0.tv_usec / 1000);
            if (delta < 1000 / fps) {
                select(1, NULL, NULL, NULL, &tv);
            } else {
                break;
            }
        }
    }

    free(raw_data_);
    free(yuv_data_);
    free(h264_data_);
    raw_data_ = NULL;
    yuv_data_ = NULL;
    h264_data_ = NULL;
}

void X264Encoder::Stop() {
    running_ = false;
}
