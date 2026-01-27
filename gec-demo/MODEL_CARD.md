# Model Card — Grammar Error Correction (GEC) ONNX Model

## 📌 Overview

This model is a fine-tuned sequence-to-sequence (Seq2Seq) transformer for English grammar correction.  
It is based on the `t5-base` architecture and trained to transform grammatically incorrect input text into corrected output.

After training, the model was optimized and exported to ONNX format for high-performance inference in a containerized environment.

This model is used as the core inference engine for the GEC Demo service.

---

## 🧠 Model Architecture

| Property        | Value |
|-----------------|--------|
| Base Model      | t5-base |
| Architecture    | Encoder–Decoder (Seq2Seq Transformer) |
| Framework       | Hugging Face Transformers |
| Format          | ONNX |
| Tokenizer       | SentencePiece |
| Language        | English |

### Exported Files

```
encoder_model.onnx
decoder_model.onnx
decoder_with_past_model.onnx
```

The `decoder_with_past_model.onnx` variant is used to accelerate autoregressive decoding.

---

## 📊 Training Data

### Dataset

- Size: ~32,000 samples
- Format:

```json
{
  "text": "incorrect sentence",
  "correct": "corrected sentence"
}
```

### Data Source

All samples were manually curated and verified using grammar correction tools and human review.
This process emphasized correctness, consistency, and coverage of common grammatical patterns.

No personal or sensitive data was intentionally included.

---

## ⚙️ Training Procedure

### Framework

* Hugging Face Transformers
* PyTorch backend

### Hardware

* CUDA-enabled GPU

### Fine-Tuning Method

* Initialized from `t5-base`
* Supervised fine-tuning on paired input/output text
* Standard Seq2Seq training objective (cross-entropy loss)

### Optimization

* Model weights converted to ONNX
* Graph optimization applied
* Decoder-with-past variant generated for faster decoding

---

## 📈 Evaluation

### Metrics

The model was evaluated using:

* BLEU score
* Exact-match accuracy (percentage of outputs identical to reference)

### Results Summary

* High accuracy on short and medium-length texts
* Strong performance on:

  * Verb tense errors
  * Subject–verb agreement
  * Article usage
  * Punctuation
  * Capitalization
  * Homophones
  * Commas
  * Apostrophes
* Reduced accuracy on very long passages

Manual review confirmed consistent improvements over raw input text.

---

## 🎯 Intended Use

This model is designed for:

* Grammar correction
* Writing assistance
* Educational tools
* Proofreading services
* Portfolio demonstration

It is intended for general English text and informal to semi-formal writing.

---

## ⚠️ Limitations

### Known Limitations

* Struggles with special Unicode characters (e.g., accented letters)
* Reduced accuracy on very long documents
* Limited understanding of slang or creative writing
* No explicit multilingual support

### Bias and Fairness

The training data reflects common English usage patterns and grammatical mistakes.
It may underrepresent certain dialects or styles.

---

## 🔒 Ethical Considerations

* Training data was manually curated
* No known personal data was intentionally included
* Model outputs should be reviewed before use in critical contexts

Users remain responsible for validating generated content.

---

## Citation

```
@article{cole_guion_2026,
      title={gec-model}, 
      author={Cole Guion},
      year={2026},
}
```
