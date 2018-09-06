
CXX=		g++
CFLAGS=		-Wall -o3 -std=c++11

COMMSRC=	src/IEncoder.cpp
X264SRC=	src/X264Encoder.cpp
GST010SRC=	src/GST010Encoder.cpp
EXP=		examples/main.cpp

X264INC=	$(shell pkg-config --cflags x264) \
		$(shell pkg-config --cflags opencv)

GST010INC=	$(shell pkg-config --cflags gstreamer-0.10) \
		$(shell pkg-config --cflags gstreamer-app-0.10)

X264LIB=	$(shell pkg-config --libs x264) \
		$(shell pkg-config --libs opencv)

GST010LIB=	$(shell pkg-config --libs gstreamer-0.10) \
		$(shell pkg-config --libs gstreamer-app-0.10)

X264OBJ=	$(COMMSRC:%.cpp=%.o) $(X264SRC:%.cpp=%.o)
GST010OBJ=	$(COMMSRC:%.cpp=%.o) $(GST010SRC:%.cpp=%.o)

all: x264 gst010

src/%.o: src/%.cpp
	$(CXX) $(CFLAGS) -c $< -o $@ $(INC)

x264: INC=$(X264INC)
x264: $(X264OBJ)
	mkdir -p build/lib
	ar -r libmenc-x264.a $(X264OBJ)
	mv libmenc-x264.a build/lib/libmenc-x264.a
	mkdir -p build/bin
	$(CXX) $(CFLAGS) $(EXP) -DUSE_X264=on -o menc-x264 -Isrc $(INC) -Lbuild/lib -lmenc-x264 $(X264LIB)
	mv menc-x264 build/bin/menc-x264

gst010: INC=$(GST010INC)
gst010: $(GST010OBJ)
	mkdir -p build/lib
	ar -r libmenc-gst010.a $(GST010OBJ)
	mv libmenc-gst010.a build/lib/libmenc-gst010.a
	mkdir -p build/bin
	$(CXX) $(CFLAGS) $(EXP) -DUSE_GST010=on -o menc-gst010 -Isrc $(INC) -Lbuild/lib -lmenc-gst010 $(GST010LIB)
	mv menc-gst010 build/bin/menc-gst010

clean:
	rm -rf $(X264OBJ)
	rm -rf $(GST010OBJ)
	rm -rf build
