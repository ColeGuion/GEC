#include "logger.h"

// (0)Critical, (1)Error, (2)Warning, (3)Info, (4)Debug
LogLev LOG_LEVEL = DEBUG;
void Log(LogLev lg, const char* format, ...)
{
    if (LOG_LEVEL >= lg)
    {
        va_list args; // List to hold the variable arguments

        if (lg <= ERROR)
            printf("\x1b[1;31mERROR: ");
        if (lg == WARNING)
            printf("\x1b[31mWARNING: ");

        // Initialize with the format string & Handle the args, then clean up the list
        va_start(args, format);
        vprintf(format, args);
        if (lg <= WARNING)
            printf("\x1b[0m");
        printf("\n");
        va_end(args);
        fflush(stdout); // Force flush the buffer
    }
}