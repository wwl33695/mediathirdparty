# DGMF Release Note
## DGMF 1.0.0
### MServer
- 加入RTSP Proxy模式
- 加入RTMP Proxy模式
- 增加Live模式下断线重连机制
- 修改Live模式下崩溃，假死等问题
- RTSP与UDP两种输入模式重构，以支持Proxy输入
- 修改UDP Input中帧率计算方式
- 修改RTSP协议栈，能够介入包含海康威视，大华，汉邦高科等厂商的RTSP

### MEnc
- 由TK1下基于GST的拉模式，改为推模式，以降低延时
- 解决帧率计算问题
- MEnc1.0.0版本后暂停维护，编码工作由新的编解码框架MCodec负责
