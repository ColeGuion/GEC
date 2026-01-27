# src/native/gec_runtime/Makefile
# Builds static library: build/libgec.a
#
# Usage:
#   From repo root:
#     make -C src/native/gec_runtime
#
# Output:
#   src/native/gec_runtime/build/libgec.a

CC       := gcc
CXX      := g++
AR       := ar

# Paths
SRC_DIR      := src
INCLUDE_DIR  := include
THIRD_DIR    := third_party
BUILD_DIR    := build

# Flags
CFLAGS   := -O3 -fPIC -I$(INCLUDE_DIR) -I$(THIRD_DIR) -I$(THIRD_DIR)/sentencepiece/include -I$(THIRD_DIR)/onnxruntime/include
CXXFLAGS := -O3 -fPIC -I$(INCLUDE_DIR) -I$(THIRD_DIR) -I$(THIRD_DIR)/sentencepiece/include -I$(THIRD_DIR)/onnxruntime/include -std=c++17

# Sources (relative to this Makefile)
C_SRCS   := $(SRC_DIR)/inference.c $(SRC_DIR)/logger.c
CPP_SRCS := $(SRC_DIR)/sentencepiece_wrapper.cpp 

# Objects go into build/ preserving subdirs
C_OBJS   := $(patsubst $(SRC_DIR)/%.c,$(BUILD_DIR)/%.o,$(C_SRCS))
CPP_OBJS := $(patsubst $(SRC_DIR)/%.cpp,$(BUILD_DIR)/%.o,$(CPP_SRCS))

LIB := $(BUILD_DIR)/libgec.a

.PHONY: all clean dirs

all: $(LIB)

dirs:
	@mkdir -p $(BUILD_DIR)

$(LIB): dirs $(C_OBJS) $(CPP_OBJS)
	$(AR) rcs $@ $(C_OBJS) $(CPP_OBJS)

# Compile C -> build/*.o
$(BUILD_DIR)/%.o: $(SRC_DIR)/%.c | dirs
	$(CC) $(CFLAGS) -c $< -o $@

# Compile C++ -> build/*.o
$(BUILD_DIR)/%.o: $(SRC_DIR)/%.cpp | dirs
	$(CXX) $(CXXFLAGS) -c $< -o $@

clean:
	rm -rf $(BUILD_DIR)
