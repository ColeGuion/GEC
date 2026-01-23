#ifndef LOGGER_H
#define LOGGER_H
#include <stdarg.h>
#include <stdio.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    CRITICAL = 0,
    ERROR = 1,
    WARNING = 2,
    INFO = 3,
    DEBUG = 4
} LogLev;

extern LogLev LOG_LEVEL;

// Updated macro to automatically pass __FILE__ and __LINE__
#define Log(lg, ...) LogMe(lg, __FILE__, __LINE__, __VA_ARGS__)

// Function used by macro
void LogMe(LogLev lg, const char* file, int line, const char* format, ...);

#ifdef __cplusplus
}
#endif

#endif // LOGGER_H