# Deep Glint Media Framework
(Deep Glint Media Framework)DGMF，是一套多媒体开发组件。旗下拥有MServer（媒体服务器），MCodec（编解码库），MCluster（集群服务器），MSpout（平台对接模块），MPlayer（媒体播放器），MMonitor（实时监测工具），MTest（辅助测试工具集）。

## MServer
MServer是一个基于Go语言开发的流媒体服务器。可以接入多种协议的流媒体数据，如rtsp, GB28182, udp等。在流媒体服务内部进行多种协议的转换，并输出为多种流媒体协议，如rtsp, rtmp, GB28181等。并且支持live, vod, proxy等多种模式。MServer提供了一组RESTful API，可以进行动态配置和状态监控

## MCodec
MCodec是一个基于C语言开发的多平台的h.264/h.265的编解码库，并支持CPU/GPU等多种编解码模式，提供动态链接库，静态链接库等多种形式，方便进行外部的计算模块调用。

## MCluster
MCluster是一个基于Go语言开发的集群服务器，提供流媒体服务的集群方案。可以进行动态扩容，媒体数据负载均衡，动态管理MServer，MSpout，MCodec等功能。MCluster提供了一组RESTful API，可以进行动态配置和状态监控

## MSpout
MSpout是一个基于C语言开发的第三方平台接入模块，可以进行如海康威视，大华，英飞拓，东方网力等使用第三方SDK进行视频数据输入的对接

## MPlayer
MPlayer是一个基于ActionScript语言开发的用来播放基于DMI协议的播放器，可以进行实时的前端可视化渲染，并提供低延时的视频播放解决方案

## MMonitor
MMonitor是一个基于Go语言开发的监测MServer媒体数据状态，监测MCluster集群状态的工具，用来分析数据源，媒体解析是否正常工作

## MTest
MTest是一个基于Go/C语言开发的辅助测试工具集，对多种数据源进行媒体协议检测（如RTSP交互过程检查，GB28181交互过程检查等），媒体信息检测（如分辨率，帧率，编码信息等），图像质量评估等