#include "sentencepiece_wrapper.h"
#include <sentencepiece_processor.h>
#include <algorithm>
#include <cstdint>
#include <cstring>
#include <string>
#include <iostream>
#include <stdint.h>
#include <fstream>
#include <vector>
#include <memory>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "wp_tokenizer.h"
#include "logger.h"
using namespace std;

void* initialize_processor(const char* model_path) {
    if (model_path == NULL) {
        return NULL;    // We can't actually process anything if we were given an invalid path
    } 
    auto* processor = new sentencepiece::SentencePieceProcessor();
    if (processor == NULL) {
        return NULL;    // Allocation failed
    }
    processor->Load(model_path);
    return static_cast<void*>(processor);
}

TokenizedTexts* prepare_texts(void* processor_ptr, char** texts, int num_texts) {
    sentencepiece::SentencePieceProcessor* processor = static_cast<sentencepiece::SentencePieceProcessor*>(processor_ptr);
    if (processor == NULL){
        Log(ERROR, "failed casting pointer to sentencepiece processor"); 
        return nullptr;
    }
    TokenizedTexts* output = new TokenizedTexts();

    // Group texts into sequences of tokens less than MAX_TOKENS
    int max_length = 0;     // Length of longest sequence
    int running_total = 0;  // # of tokens in current_group
    std::vector<std::vector<int>> grouped_ids = {};
    std::vector<int> current_group = {};
    std::vector<std::string> newlineStrings = {};

    for (int i = 0; i < num_texts; ++i) {
        // Check for newline characters
        if (strchr(texts[i], '\n') != nullptr) {
            // Append group if not empty
            if (!current_group.empty()) {
                grouped_ids.push_back(current_group);
                
                // Update max_length
                if (running_total > max_length) max_length = running_total;
                
                // Clear the current_group
                current_group = {};
                running_total = 0;
            }

            // Append newline strings to their own seperate groups
            int newLn_idx = (int)grouped_ids.size() + (int)newlineStrings.size();
            output->newline_inds[(int)newlineStrings.size()] = newLn_idx;
            newlineStrings.push_back(texts[i]);
        } else {
            // Check if we've reached the maximum number of texts
            if (grouped_ids.size() >= MAX_BATCH_SIZE) {
                break;
            }
            // Encode the text into token IDs
            std::vector<int> pieces;
            processor->Encode(texts[i], &pieces);
            int pieceSz = (int)pieces.size();

            // Make the max number of new tokens no more than 20 tokens more than the original sequences
            if ((running_total+pieceSz) > MAX_TOKENS-20 && !current_group.empty()) {
                // Current group cannot fit more tokens so append it to grouped_ids
                grouped_ids.push_back(current_group);
                
                // Update max_length
                if (running_total > max_length) max_length = running_total;

                // Clear the current_group
                current_group = pieces;
                running_total = pieceSz;
            } else {
                // Append ids to the end of current_group using insert
                current_group.insert(current_group.end(), pieces.begin(), pieces.end());
                running_total += pieceSz;
            }
        }
    }
    if (!current_group.empty() && grouped_ids.size() < MAX_BATCH_SIZE) {
        // Append remaining group if any
        grouped_ids.push_back(current_group);
        if (running_total > max_length) max_length = running_total;
    }

    // Max Length of sequences
    max_length += 1;    // Add 1 for the EOS ("</s>") token
    if (max_length > MAX_TOKENS) {
        Log(DEBUG, "Max length is too big(%d). Resetting down to MAX_TOKENS(%d)", max_length, MAX_TOKENS);
        max_length = MAX_TOKENS;
    } 
    
    // Set output variables
    int num_tokens = (int)grouped_ids.size() * max_length;   // Total number of tokens
    output->shape[0] = grouped_ids.size();
    output->shape[1] = max_length;
    output->data_len = num_tokens * sizeof(int64_t);
    output->newline_size = (int)newlineStrings.size();

    // Set the token ids and attention mask
    for (int i = 0; i < (int)output->shape[0]; ++i) {
        int pieces_size = (int)grouped_ids[i].size();

        // Iterate through the pieces in each group and convert them to token ids
        for (int j=0; j<max_length; ++j) {
            int idx = (i*max_length) + j;    // Index of the token in the array
            if (j < pieces_size) {
                output->attention_mask[idx] = 1;
                output->ids[idx] = grouped_ids[i][j];
                
                // IF the sequence has reached the maximum length, cap it off with an EOS token
                if (j+1 == max_length) {
                    output->ids[idx] = 1;
                }
            } else if (j == pieces_size) {
                // Add the </s> token id to the end of the sequence
                output->attention_mask[idx] = 1;
                output->ids[idx] = 1;
            } else {
                // Add padding tokens
                output->attention_mask[idx] = 0;
                output->ids[idx] = 0;
            }
        }
    }

    // Convert vector of newline strings to a char** array
    output->newline_strs = new char*[newlineStrings.size() + 1];
    for (int i = 0; i < (int)newlineStrings.size(); i++) {
        output->newline_strs[i] = strdup(newlineStrings[i].c_str());
    }
    // NULL terminate the array
    output->newline_strs[newlineStrings.size()] = nullptr;
    
    return output;
}

char* decode_texts(void* processor_ptr, int decoded_ids[MAX_BATCH_SIZE][MAX_TOKENS], TokenizedTexts* tokensObj) {
    std::string final_text = "";
    int num_texts = (int)tokensObj->shape[0];
    sentencepiece::SentencePieceProcessor* processor = static_cast<sentencepiece::SentencePieceProcessor*>(processor_ptr);
    if (processor == NULL){
        Log(ERROR, "failed casting pointer to sentencepiece processor");  //std::cerr << "failed casting pointer to sentencepiece processor\n";
        return nullptr;
    }

    // Decode the token IDs into texts
    int true_idx = 0;   // Tracks index of texts including newline strings
    int newLn_count = 0;
    for (int i = 0; i < num_texts; ++i) {
        if (newLn_count < tokensObj->newline_size && true_idx == tokensObj->newline_inds[newLn_count]) {
            // Previous text was not a newline literal so remove the space at the end
            if (final_text != "") final_text.pop_back();

            // Append newline string to the final text
            final_text += tokensObj->newline_strs[newLn_count];
            newLn_count++;
            true_idx++;
        }
        
        // Convert int* array to std::vector<int> & Remove unknown token IDs(2)
        std::vector<int> dec_ids(decoded_ids[i], decoded_ids[i] + MAX_TOKENS);
        dec_ids.erase(std::remove(dec_ids.begin(), dec_ids.end(), 2), dec_ids.end());
        
        // Decode the token IDs
        std::string res;
        processor->Decode(dec_ids, &res);
        
        // Add a space between texts
        final_text += res + " ";
        if (i == num_texts-1) {
            final_text.pop_back();  // Remove the last space
        }
        true_idx++;
    }
    if (newLn_count < tokensObj->newline_size && true_idx == tokensObj->newline_inds[newLn_count]) {
        // Append newline string to the final text
        final_text += tokensObj->newline_strs[newLn_count];
    }


    // Convert to a char* string
    //char* result = strdup(final_text.c_str());
    // Allocate memory for the C-style string
    char* result = (char*)malloc(final_text.length() + 1);  // +1 for the null terminator
    if (result == nullptr) {
        std::cerr << "Memory allocation failed." << std::endl;
        return nullptr;
    }

    // Copy the contents of the std::string to the new char*
    std::strcpy(result, final_text.c_str());
    return result;
}

Tokenized_WP_Output batch_gibb_texts(const char* gibb_tok_path, char** texts, int batchSize) {
    WordPieceTokenizer wp_tokenizer(gibb_tok_path);   // Instantiate the tokenizer (Auto freed when it goes out of scope)
    Tokenized_WP_Output output;
    std::vector<std::vector<size_t>> batch_input_ids(batchSize);    // Holds the encoded input IDs for each sequence
    int maxLen = 0; // Max length of a sequence in the batch

    for (int i=0; i<batchSize; i++) {
        // Lowercase the text
        std::string input_text = texts[i];
        std::transform(input_text.begin(), input_text.end(), input_text.begin(), ::tolower);

        // Tokenize the input to get input IDs
        std::vector<size_t> input_ids = wp_tokenizer.tokenize_full(utf8_to_wstring(input_text));
        size_t num_tokens = input_ids.size();
        if ((size_t)maxLen < num_tokens) {
            maxLen = num_tokens;
        }

        // Append vector to batch_input_ids
        batch_input_ids[i] = std::move(input_ids); // Move to avoid copy
    }

    // Set output length
    output.shape[0] = std::min(batchSize, MAX_BATCH_SIZE);
    output.shape[1] = std::min(maxLen, MAX_TOKENS);
    output.length = output.shape[0] * output.shape[1];

    int index = 0;
    // Add the token IDs to the IDs and attention_mask arrays
    for (int i=0; i<(int)output.shape[0]; i++) {
        for (int j=0; j<(int)output.shape[1]; j++) {
            if (j >= (int)batch_input_ids[i].size()) {
                // Padding
                output.ids[index] = 0;
                output.attention_mask[index] = 0;
            } else {
                output.ids[index] = batch_input_ids[i][j];
                output.attention_mask[index] = 1;
            }
            index++;
        }
    }

    return output;
}

// Free the SentencePieceProcessor instance
void free_processor(void* processor_ptr) {
    if (processor_ptr != nullptr) {
        auto* processor = static_cast<sentencepiece::SentencePieceProcessor*>(processor_ptr);
        delete processor;
    }
}

// Free the TokenizedTexts objects
void free_tokenized_texts(TokenizedTexts* obj) {
    if (obj != nullptr) {
        if (obj->newline_strs != nullptr) {
            for (size_t i = 0; obj->newline_strs[i] != nullptr; ++i) {
                free(obj->newline_strs[i]);
            }
            delete[] obj->newline_strs;
            //delete static_cast<std::vector<std::string>*>(obj->newline_strs);
            //obj->newline_strs = nullptr;
        }
        delete obj;
    }
}

// Free allocated memory for Tokenized_WP_Output
void free_tokenized_wp_output(Tokenized_WP_Output output) {
    // Nothing to free
}
