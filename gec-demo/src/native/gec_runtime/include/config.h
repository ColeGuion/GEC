// config.h
// Defines constants for other files to share and reference
#ifndef CONFIG_H
#define CONFIG_H

#define LOGIT_SIZE 32128    // Logit Tensor Shape = BatchSize x 1 x 32128
#define GIBB_CLASSES 4      // Clean, Mild, Word-Salad, Noise
#define MAX_TOKENS 100      // Maximum sequence length allowed (NOTE: No safety bounds are in place to enforce or set this limit) (Maybe prepare_texts function should be updated to handle this)
#define MAX_BATCH_SIZE 500  // Maximum batch size allowed (NOTE: No safety bounds are in place to enforce or set this limit)

#endif // CONFIG_H