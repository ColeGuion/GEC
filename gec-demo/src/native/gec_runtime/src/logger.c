#include "logger.h"
#include <string.h>

// (0)Critical, (1)Error, (2)Warning, (3)Info, (4)Debug
LogLev LOG_LEVEL = DEBUG;

void LogMe(LogLev lg, const char* file, int line, const char* format, ...) {
    if (LOG_LEVEL >= lg) {
        va_list args; // List to hold the variable arguments
        
        // Extract just the filename from the full path
        const char* filename = strrchr(file, '/');
        if (filename == NULL) {
            filename = file;  // No path separator found, use the whole string
        } else {
            filename++;  // Skip the '/'
        }
        
        // Print log level with color and filename/line number
        if (lg <= CRITICAL) {
            printf("\x1b[91m[CRITICAL] %s:%d: ", filename, line);
        } else if (lg == ERROR) {
            printf("\x1b[91m[ERROR] %s:%d: ", filename, line);
        } else if (lg == WARNING) {
            printf("\x1b[31m[WARNING] %s:%d: ", filename, line);
        } else if (lg == INFO) {
            printf("[INFO] %s:%d: ", filename, line);
        } else if (lg == DEBUG) {
            printf("[DEBUG] %s:%d: ", filename, line);
        }
        
        // Print the actual message
        va_start(args, format);
        vprintf(format, args);
        va_end(args);
        
        // Reset color and add newline
        printf("\x1b[0m\n");
        fflush(stdout);
    }
}