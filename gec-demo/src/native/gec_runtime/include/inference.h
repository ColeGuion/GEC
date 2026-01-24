#ifndef INFERENCE_H
#define INFERENCE_H

#include <assert.h>
#include <malloc.h>
#include <math.h>
#include <stdarg.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
#include "config.h"
#include "logger.h"
#include "onnxruntime_c_api.h"

// Decoder Input/Output Names
extern char* decoder_output_names[51];
extern char* decPast_input_names[51];
extern char* decPast_output_names[25];

// G.E.C.O. => Grammar Error Corrector Onnx
typedef struct {
    OrtValue* input_tensor;
    OrtValue* output_tensor;
    OrtValue* output_tensor_fp16;
    OrtValue** binded_tensors; // Array of tensors to be filled with output tensors in decoded
    size_t binded_tensors_len;

    const OrtApi* g_ort;
    OrtEnv* env;
    OrtMemoryInfo* memory_info;
    OrtMemoryInfo* cuda_memory_info;
    OrtRunOptions* run_options;
    OrtSessionOptions* session_options;
    OrtArenaCfg* arena_cfg;
    char device_id[8];

    // Model Sessions
    OrtSession* encoder_session;
    OrtSession* decoder_session;
    OrtSession* decPast_session;
    OrtSession* gibb_session;

    // Allocators
    OrtAllocator* allocator;

    // IO Bindings
    OrtIoBinding* enc_io_binding;
    OrtIoBinding* dec_io_binding;
    OrtIoBinding* decPast_io_binding;
    OrtIoBinding* gibb_io_binding;

    // Generated Tokens
    int generated_tokens[MAX_BATCH_SIZE][MAX_TOKENS]; // Array of generated tokens for each sequence in the batch

    // SentencePiece Utilities
    void* processor;
} Geco;


// Goto the cleanup label if an error occurs
#define ORT_CLEAN_ON_ERROR(cleanup_label, geco, expr)                                              \
    do {                                                                                           \
        OrtStatus* onnx_status = (expr);                                                           \
        if (onnx_status != NULL) {                                                                 \
            const char* msg = geco->g_ort->GetErrorMessage(onnx_status);                           \
            fprintf(stderr, "[%d] %s() - ERROR: %s\n", __LINE__, __func__, msg);                   \
            geco->g_ort->ReleaseStatus(onnx_status);                                               \
            goto cleanup_label;                                                                    \
        }                                                                                          \
        geco->g_ort->ReleaseStatus(onnx_status);                                                   \
    } while (0);


/**
 * @brief Initialize a new GECO instance
 *
 * @param useGpu Boolean to determine if the GPU should be used
 * @param gpuId ID of the gpu to use
 */
void* NewGeco(int log_level, bool use_gpu, int gpu_id);

/**
 * @brief Frees all of the allocated memory used in this GECO object
 *
 * @param objPtr Context for a GECO object
 */
void FreeGeco(void* objPtr);

// Release the input tensor
void free_inpTensor(Geco* geco);
void free_binded_tensors(Geco* geco);


/**
 * @brief Checks if last 6 tokens in an array are the same or alternating between two distinct
 * values.
 *
 * @param arr Array of tokens in a sequence to be checked for repeating tokens.
 * @param lastInd The last index with a valid value in the array. (Not a padding 0 token)
 *
 * @return True if the last 6 tokens are repeating/alternating. Returns False otherwise
 */
bool checkRepeating(int* arr, int lastInd);

/**
 * @brief Takes a logits tensor and adds the maximum token IDs for each sequence into the newTokens
 * array.
 *
 * @param geco GECO object for context
 * @param newTokens Array of batchSize-many values to hold the token ID of the most probable next
 * token for each sequence
 * @param batchSize Number of texts being processed at once
 * @param completed_sequences Array marking which sequences have already been completed
 * @param runNum Number of run
 *
 * @return 0 if successful, -1 if an error occurs
 */
int getMaxTokens(Geco* geco, int64_t* newTokens, int batchSize, int* completed_sequences, int runNum);


/**
 * @brief Recursive run of decoder_wtih_past_model.onnx
 * Start by running the model and return if all the sequences are complete
 * Otherwise, reshape and bind new input and output tensors based on the previous outputs and run
 * again
 *
 * @param geco GECO objects containing the ONNX Runtime API and other necessary objects.
 * @param runNum Iteration of recursive run. Helpful for tracking generated tokens at a given run.
 * @param nextToks Array holding the next generated tokens from a run. They are used as input to the
 * model.
 * @param batchSize Number of texts being processed at once.
 * @param completed_sequences Array marking if a sequence has completed so we can stop searching for
 * its next tokens.
 */
void runPast(Geco* geco, int runNum, int64_t* nextToks, int batchSize, int* completed_sequences);

/**
 * @brief Binds the input and output tensors from the decoder model to get ready for the next run
 * Then it will call runPast() and recursively run on past outputs until all sequences are complete
 *
 * @param geco GECO objects containing the ONNX Runtime API and other necessary objects.
 * @param lastHiddenState Last Hidden State tensor to be used as input to the model.
 * @param Ort_AttnMask Attention Mask tensor to be used as input to the model.
 * @param batchSize Number of texts being processed at once
 *
 * @return OrtValue** The outputs from the decoder model, which will be used as inputs to the
 * decoder_with_past model.
 */
void runDecoders(Geco* geco, int batchSize);

/**
 * @brief Function called by a GECO object to use its context for inferencing.
 * Converts void* to Geco* and runs inference in InferModel()
 *
 * @param context GECO object to run the inference with
 * @param texts Array of texts to be processed
 * @param num_texts Number of texts split into sentences
 *
 * @return String containing the corrected text.
 */
void GecoRun(void* context, char** texts, int num_texts, char** result);
void InferModel(Geco* geco, char** texts, int num_texts, char** result);

#endif // INFERENCE_H