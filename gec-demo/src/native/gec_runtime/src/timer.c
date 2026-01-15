// timer.c
#include "timer.h"

// Timer names
#define NUM_TIMERS 16
const char* time_names[NUM_TIMERS] = {
    "Load Config",
    "Create Geco",
    "GEC Preproc",
    "Encoder",
    "Main Decoder",
    "Past Decoder",
    "All Decoders",
    "Get Max Tokens",
    "GEC Postproc",
    "GEC Total",
    "Gibb Preproc",
    "Gibb Model Run",
    "Calculate Softmax",
    "Gibberish Total",
    "Destroy Geco",
    "Session Run",
};
double allTimes[NUM_TIMERS] = {0};


void SetTimerValue(const char* name, clock_t start, clock_t end) {
    double total_time = ((double)(end - start)) / CLOCKS_PER_SEC;
    total_time = round(total_time * 1000.0) / 1000.0; // Get to the third decimal place

    for (int i = 0; i < NUM_TIMERS; i++) {
        if (strcmp(time_names[i], name) == 0) {
            allTimes[i] += total_time;
            return;
        }
    }
    Log(DEBUG, "Invalid timer name: %s", name);
}

void SetGpuTime(struct timespec start) {
    struct timespec end;
    clock_gettime(CLOCK_MONOTONIC, &end);
    double total_time = (end.tv_sec - start.tv_sec) + (end.tv_nsec - start.tv_nsec) / 1e9;
    allTimes[15] += total_time;
}

// Print and clear the results from the set timed values
void PrintTimes() {
    Log(DEBUG, "Timed Results:");
    for (int i = 0; i < NUM_TIMERS; i++) {
        if (allTimes[i] != 0) {
            Log(DEBUG, "   %.3f - %s", allTimes[i], time_names[i]);
            allTimes[i] = 0;
        }
    }
}
