# src/gec/Makefile
# Builds `libgec.a` from `src/gec/`
# From `src/` you can build with this command: `make -C gec`

CC      := gcc
CXX     := g++
AR      := ar
CFLAGS  := -O3 -fPIC -I.
CXXFLAGS:= -O3 -fPIC -I. -std=c++17

# Add any extra include paths you need here, e.g. sentencepiece headers if external.
# CXXFLAGS += -I/path/to/sentencepiece/include

C_SRCS   := inference.c logger.c timer.c
CPP_SRCS := wp_tokenizer.cpp sentencepiece_wrapper.cpp

C_OBJS   := $(C_SRCS:.c=.o)
CPP_OBJS := $(CPP_SRCS:.cpp=.o)

LIB := libgec.a

all: $(LIB)

$(LIB): $(C_OBJS) $(CPP_OBJS)
	$(AR) rcs $@ $^

%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@

%.o: %.cpp
	$(CXX) $(CXXFLAGS) -c $< -o $@

clean:
	rm -f $(C_OBJS) $(CPP_OBJS) $(LIB)
