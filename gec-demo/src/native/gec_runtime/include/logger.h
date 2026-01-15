#ifndef LOGGER_H
#define LOGGER_H
#include <stdarg.h>
#include <stdio.h>

#ifdef __cplusplus
extern "C"
{
#endif

    // Function to log messages
    typedef enum
    {
        CRITICAL,
        ERROR,
        WARNING,
        INFO,
        DEBUG
    } LogLev;
    extern LogLev LOG_LEVEL;
    void Log(LogLev lg, const char* format, ...);

#ifdef __cplusplus
}
#endif

#endif // LOGGER_H