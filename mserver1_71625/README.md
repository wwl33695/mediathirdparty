# MServer

## Overview

MServer is a versatile streaming media server, which is responsible for data conversion based on H.264 ES byte stream. Currently, the latest version is 0.8.x. This version contains a streaming media server based on RTSP protocol and a control server based on HTTP protocol.You can watch live video through a media player such as VLC, and you can check the current state of the server through a set of RESTful style interface, and control it. In addition, you can configure the data source in a variety of formats, such as: RTSP, UDP, and file. And the configuration of the process is dynamic, in other words, you do not need to restart the service to finish the changes. All the code for this project is completed by the Go language. You can easily compile and cross compile the project on the platform that supports the Go language.

## How to Build
```
$ cd mserver
$ ./make.sh
```
or

```
$ cd mserver
$ go build
```
## How to Run
```
$ ./mserver-[OS]_[Architecture] [-c config path] (config.json by default)

$ ./mserver-darwin_x86_64 (on MacOSX)
$ ./mserver-linux_armv7l (on nvidia jetson tk1)
$ ./mserver-linux_x86_64 (on linux x86)
```
## Config File
```
{
   "RTSPServer":{
      "Enable":true,
      "Port":8554,
      "Param":null
   },
   "HTTPAPIServer":{
      "Enable":true,
      "Port":8080,
      "Param":null
   },
   "RTMPServer":{
      "Enable":true,
      "Port":1935,
      "Param":null
   },
   "LiveInputs":{
      "test0":{
         "StreamId":"test0",
         "Uri":"udp://127.0.0.1:9002",
         "Protocols":{
            "rtmp":{
               "Protocol":"rtmp",
               "Enable":true
            },
            "rtsp":{
               "Protocol":"rtsp",
               "Enable":true
            }
         }
      },
      "test1":{
         "StreamId":"test1",
         "Uri":"rtsp://admin:deepglint123@192.168.4.111:554/h264/ch1/main/av_stream",
         "Protocols":{
            "rtmp":{
               "Protocol":"rtmp",
               "Enable":true
            },
            "rtsp":{
               "Protocol":"rtsp",
               "Enable":true
            }
         }
      },
      "test2":{
         "StreamId":"test2",
         "Uri":"file:///Users/Leo/Desktop/LT/test2.h264",
         "Protocols":{
            "rtmp":{
               "Protocol":"rtmp",
               "Enable":true
            },
            "rtsp":{
               "Protocol":"rtsp",
               "Enable":true
            }
         }
      }
   },
   "VodInputs":{

   }
}
```
## Inputs
### Live UDP

+ URL

```
udp://127.0.0.1:[port]
```
+ Format

H.264 Element Byte Stream over udp, one H.264 Nalu as a packet. It must have start code 0x00, 0x00, 0x00, 0x01. Mserver can get resolution and other media information from SPS and PPS frame
### Live RTSP
+ URL

```
rtsp://[ip]:[port]/[path]
```
+ Format

RTSP over TCP or UDP
### Live H.264 File
+ URL

```
file://[path].h264
```
+ Format

H.264 Element Byte Stream in file. It must have start code 0x00, 0x00, 0x00, 0x01. Mserver can get resolution and other media information from SPS and PPS frame

### Reference
https://www.ietf.org/rfc/rfc2326.txt
## HTTP API Server
### Status API
```
Request:
GET /status

Response:
{
  "StreamCount": 1,
  "SessionCount": 2,
  "LiveStreams": [
    {
      "StreamId": "test0",
      "InputStatus": true,
      "URI": "udp://127.0.0.1:9002",
      "Fps": 25,
      "Index": 807,
      "Width": 640,
      "Height": 480,
      "SPS": "Z0LAHtoCgPaEAAADAAQAAAMAyjxYuoA=",
      "PPS": "aM4PyA==",
      "SessionsStatus": [
        {
          "SessionId": "29819e8c-c94b-4d8f-94d1-c548688dfb1d",
          "RemoteAddr": "127.0.0.1:59359",
          "Network": "tcp",
          "Protocol": "rtsp"
        },
        {
          "SessionId": "6103a049-abdc-4c70-bd3c-8098dee6e894",
          "RemoteAddr": "127.0.0.1:59360",
          "Network": "tcp",
          "Protocol": "rtmp"
        }
      ],
      "Protocols": {
        "rtmp": true,
        "rtsp": true
      }
    }
  ],
  "VodStreams": null,
  "MediaServers": [
    {
      "Protocol": "rtsp",
      "Enable": true,
      "Port": 8554,
      "Param": null
    },
    {
      "Protocol": "rtmp",
      "Enable": true,
      "Port": 1935,
      "Param": null
    }
  ]
}
```

### Version API
```
Request:
GET /version

Response:
MServer/0.9.0 (Deep Glint Inc. 2016.08.31)
```

### Config API
```
Request:
GET /config

Response:
[json format of config file]
```

### Add Input API
```
Request:
GET /add-input?stream_id=test0&stream_type=live&uri=udp://127.0.0.1:9004

Response:
{
  "StatusCode": 200,
  "Content": "OK"
}
```

### Remove Input API
```
Request:
GET /remove-input?stream_id=0&stream_type=live

Response:
{
  "StatusCode": 200,
  "Content": "OK"
}
```

### Add Output API
```
Request:
GET /add-output?stream_id=test0&stream_type=live&protocol=rtsp

Response:
{
  "StatusCode": 200,
  "Content": "OK"
}
```

### Add GB28181_Output API
```
Request:
POST URL: /add-output?stream_id=gb28181&stream_type=live&protocol=gb28181

Body:
{
    "ID": "34020000001320000001",
    "Name": "firtst channel",
    "Manufacturer": "DEEPGLINT",
    "Model": "IP Camera",
    "Owner": "Owner",
    "CivilCode": "CivilCode",
    "Address": "Address",
    "Parental": 0,
    "SafetyWay": 0,
    "RegisterWay": 1,
    "Secrecy": 0,
    "Status": "ON"
}

Response:
{
  "StatusCode": 200,
  "Content": "OK"
}
```
### Reference
http://download.csdn.net/detail/lgstudyvc/4682937
### Start Server API
```
request /start-server?serverid=rtmp&port=1935

Response:
{
  "StatusCode": 200,
  "Content": "OK"
}
```


### Start GB28181_Server API
```
request
POST URL:/start-server?serverid=gb28181&port=5060

Body:
{
    "Interval": 5,
    "DeviceID": "34020000001180000001",
    "DeviceAreaID": "3402000000",
    "ServerID": "34010000002000000001",
    "ServerAreaID": "3401000000",
    "ServerHost": "192.168.1.111",
    "ServerPort": 5060,
    "ServerPassword": "12345678"
}

Response:
{
  "StatusCode": 200,
  "Content": "OK"
}
```

### Stop Server API
```
request /stop-server?serverid=rtmp

Response:
{
  "StatusCode": 200,
  "Content": "OK"
}
```

### Remove Output API
```
Request:
GET /remove-output?streamid=test0&islive=1&protocol=rtsp

Response:
{
  "StatusCode": 200,
  "Content": "OK"
}
```

## RTSP Server
### Live
```
rtsp://[ip]:[port]/live/[stream_id]
```

### Proxy
```
rtsp://[ip]:[port]/proxy/[stream_id]?uri=[rtsp url]
```

## RTMP Server
### Live
```
rtmp://[ip]:[port]/[live]/[stream_id]
```

### Proxy
```
rtmp://[ip]:[port]/proxy/[stream_id]?uri=[rtsp url]
```