// timer.h
#ifndef TIMER_H
#define TIMER_H

#include <stdio.h>
#include <stdint.h>
#include <time.h>
#include <unistd.h>
#include <malloc.h>
#include <math.h>
#include <string.h>
#include "logger.h"

// Set the time value of a function
void SetTimerValue(const char* name, clock_t start, clock_t end);

// Set time used when running GPU
void SetGpuTime(struct timespec start);

// Print and clear the results from the set timed values
void PrintTimes();

#endif // TIMER_H