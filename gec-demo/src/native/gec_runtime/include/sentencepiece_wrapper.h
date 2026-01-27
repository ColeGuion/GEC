#ifndef SENTENCEPIECE_WRAPPER_H
#define SENTENCEPIECE_WRAPPER_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "config.h"

// Structs
typedef struct {
    int64_t ids[MAX_BATCH_SIZE * MAX_TOKENS];
    int64_t attention_mask[MAX_BATCH_SIZE * MAX_TOKENS];
    int64_t shape[2];
    size_t data_len;
    int newline_size;                 // Number of newline strings
    int newline_inds[MAX_BATCH_SIZE]; // Indicies of the newline strings in the full array of texts
    char** newline_strs;              // Newline strings to add back to the text later
} TokenizedTexts;

#ifdef __cplusplus
extern "C" {
#endif

/**
 * @brief Initializes the SentencePiece model processor
 *
 * @param model_path Path to the SentencePiece model file
 *
 * @return void* Pointer to the SentencePieceProcessor object
 */
void* initialize_processor(const char* model_path);

/**
 * @brief Groups the texts into strings less than total MAX_TOKENS
 * Then tokenizes those strings and make them padded to the same length
 *
 * @param processor_ptr Void pointer to the SentencePieceProcessor object
 * @param texts Array of strings split by sentences and newline string
 * @param num_texts Number of texts in the texts array
 *
 * @return Structure containing the tokenized ids and attention mask, as well as
 * newline string info to add back once the models results are decoded
 */
TokenizedTexts* prepare_texts(void* processor_ptr, char** texts, int num_texts);

/**
 * @brief Groups the texts into strings less than total max_tokens
 * Then tokenizes those strings
 *
 * @param processor_ptr Void pointer to the SentencePieceProcessor object
 * @param decoded_ids Array of token IDs from the decoder model which will be turned into text
 * @param tokensObj Pointer to the TokenizedTexts object
 *
 * @return The final grammatically corrected string
 */
char* decode_texts(void* processor_ptr,
                   int decoded_ids[MAX_BATCH_SIZE][MAX_TOKENS],
                   TokenizedTexts* tokensObj);

/**
 * @brief Free memory allocated by SentencePieceProcessor object
 *
 * @param processor_ptr Pointer to the SentencePieceProcessor object
 */
void free_processor(void* processor_ptr);

/**
 * @brief Free memory allocated by TokenizedTexts object
 *
 * @param obj TokenizedTexts instance
 */
void free_tokenized_texts(TokenizedTexts* obj);

#ifdef __cplusplus
}
#endif

#endif // SENTENCEPIECE_WRAPPER_H
