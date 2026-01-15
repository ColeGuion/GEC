#ifndef SENTENCEPIECE_WRAPPER_H
#define SENTENCEPIECE_WRAPPER_H

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Constants
#define LOGIT_SIZE 32128 // Logit Tensor Shape = BatchSize x 1 x 32128
#define GIBB_CLASSES 4   // Clean, Mild Gibberish, Word Salad, Noise
#define MAX_TOKENS 100      // Maximum sequence length allowed (NOTE: No safety bounds are in place to enforce or set this limit) (Maybe prepare_texts function should be updated to handle this)
#define MAX_BATCH_SIZE 500  // Maximum batch size allowed (NOTE: No safety bounds are in place to enforce or set this limit)


// Types
typedef struct {
    int64_t ids[MAX_BATCH_SIZE * MAX_TOKENS];
    int64_t attention_mask[MAX_BATCH_SIZE * MAX_TOKENS];
    int64_t shape[2];
    int length;
} Tokenized_WP_Output;

typedef struct {
    int64_t ids[MAX_BATCH_SIZE * MAX_TOKENS]; // Array of token IDs
    int64_t
        attention_mask[MAX_BATCH_SIZE * MAX_TOKENS]; // Attention mask marking non-padding tokens
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
 * @brief Tokenize a batch of texts using the WordPiece tokenizer
 *
 * @param gibb_tok_path Path to the gibberish tokenizer model
 * @param texts Array of texts to be tokenized
 * @param batchSize Number of texts in the array
 *
 * @return A structure containing the tokenized ids and attention mask and their shapes
 */
Tokenized_WP_Output batch_gibb_texts(const char* gibb_tok_path, char** texts, int batchSize);


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

/**
 * @brief Free memory allocated by Tokenized_WP_Output object
 *
 * @param output Tokenized_WP_Output instance
 */
void free_tokenized_wp_output(Tokenized_WP_Output output);

#ifdef __cplusplus
}
#endif

#endif // SENTENCEPIECE_WRAPPER_H
