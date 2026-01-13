
**TODO: Make misspellings either be the type `SPELLING_MISTAKE` or `PROFANITY`**


# Features
- Go Channels to handle multi incoming requests
- Log Levels



# What Needs Included

### `cJSON.h`
```makefile
LDFLAGS = -lcjson -lm
$(TARGET): main.c
	$(CC) $(CFLAGS) $< -o $@ $(LDFLAGS)
```

```cpp
#include <cjson/cJSON.h>
```
 




# Install cJSON 

This installs the header to `/usr/include/cJSON.h`
```bash
sudo apt update
sudo apt install libcjson-dev
```

Check where the header is installed:
```bash
dpkg -L libcjson-dev | grep cJSON.h
```



# Old CFlags
```
#cgo CFLAGS: -I./gec -I./gec/../.. -I${SRCDIR}/../gec -I/usr/local/include
#cgo CXXFLAGS: -std=c++11
#cgo LDFLAGS: -L./gec -L${SRCDIR}/../gec -linference -lstdc++ -L/usr/local/lib -ljson-c -lsentencepiece -licuuc -licudata -lonnxruntime -lonnxruntime_providers_shared -lonnxruntime_providers_cuda -lcuda -lcudar -lm

// Include your C headers
#include <stdlib.h>
#include "gec/inference.h"
#include "gec/logger.h"
#include "gec/timer.h"
```


```

// WT-AI LDFLAGS: -L. -L/usr/local/lib -ljson-c -lsentencepiece -lonnxruntime -lonnxruntime_providers_shared -lonnxruntime_providers_cuda -lcuda -lcudart -lm -lstdc++ -licuuc -licudata
// missing: -linference
/*
Include these later:
	#include <stdlib.h>
	#include "inference.h"
?	-I../gec

C Flags:
-I. -I/usr/local/include

LD Flags:
-L. -L/usr/local/lib -linference -ljson-c -lsentencepiece -lstdc++ -licuuc -licudata -lonnxruntime -lonnxruntime_providers_shared -lonnxruntime_providers_cuda -lcuda -lcudar -lm
*/

```