#ifndef WP_TOKENIZER_H
#define WP_TOKENIZER_H

#include "json.hpp"

#include <iostream>
#include <string>
#include <unicode/uchar.h> // ICU library
#include <vector>
using namespace std;
using json = nlohmann::json;

// Utility functions
std::string trim(const std::string& original);
std::vector<std::wstring> split(const std::wstring& input);
bool isPunctuation(UChar32 charCode);
bool _is_punctuation(UChar32 c);
std::vector<std::wstring> run_split_on_func(const std::wstring& text);
std::string wstring_to_utf8(const std::wstring& wstr);
std::wstring utf8_to_wstring(const std::string& str);

// Class definition for WordPieceTokenizer
class WordPieceTokenizer
{
  private:
    json jsonObj;
    json vocab;
    size_t max_input_chars_per_word;
    std::wstring unk_token;

  public:
    // Constructor
    WordPieceTokenizer(const std::string& config_path);

    // Method to retrieve the index of a word in the vocabulary
    int get_word_index(const std::wstring& word);

    // Method to tokenize an entire text input
    std::vector<size_t> tokenize_full(const std::wstring& input_text);

    // Method to tokenize a single word using wordpiece tokenization
    std::vector<std::wstring> wordpiece_tokenize(const std::wstring& input_text);

    // Convert a sequence of tokens to their respective IDs
    std::vector<size_t> convert_tokens_to_ids(const std::vector<std::wstring>& input_seq);
};

#endif // WP_TOKENIZER_H
