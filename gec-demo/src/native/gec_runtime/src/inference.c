#include "inference.h"
#include "sentencepiece_wrapper.h"

// Path Variables
static char PATH_ENCODER[1000];
static char PATH_DECODER[1000];
static char PATH_DECODER_PAST[1000];
static char PATH_GIBB_TOKENIZER[1000];
static char PATH_GIBB[1000];
static char PATH_SP_MODEL[1000];

// Decoder Input/Output Names
char* decoder_output_names[51] = {"logits", "present.0.decoder.key", "present.0.decoder.value", "present.0.encoder.key", "present.0.encoder.value", "present.1.decoder.key", "present.1.decoder.value", "present.1.encoder.key", "present.1.encoder.value", "present.2.decoder.key", "present.2.decoder.value", "present.2.encoder.key", "present.2.encoder.value", "present.3.decoder.key", "present.3.decoder.value", "present.3.encoder.key", "present.3.encoder.value", "present.4.decoder.key", "present.4.decoder.value", "present.4.encoder.key", "present.4.encoder.value", "present.5.decoder.key", "present.5.decoder.value", "present.5.encoder.key", "present.5.encoder.value", "present.6.decoder.key", "present.6.decoder.value", "present.6.encoder.key", "present.6.encoder.value", "present.7.decoder.key", "present.7.decoder.value", "present.7.encoder.key", "present.7.encoder.value", "present.8.decoder.key", "present.8.decoder.value", "present.8.encoder.key", "present.8.encoder.value", "present.9.decoder.key", "present.9.decoder.value", "present.9.encoder.key", "present.9.encoder.value", "present.10.decoder.key", "present.10.decoder.value", "present.10.encoder.key", "present.10.encoder.value", "present.11.decoder.key", "present.11.decoder.value", "present.11.encoder.key", "present.11.encoder.value"};
char* decPast_input_names[51] = {"input_ids", "encoder_attention_mask", "encoder_hidden_states", "past_key_values.0.decoder.key", "past_key_values.0.decoder.value", "past_key_values.0.encoder.key", "past_key_values.0.encoder.value", "past_key_values.1.decoder.key", "past_key_values.1.decoder.value", "past_key_values.1.encoder.key", "past_key_values.1.encoder.value", "past_key_values.2.decoder.key", "past_key_values.2.decoder.value", "past_key_values.2.encoder.key", "past_key_values.2.encoder.value", "past_key_values.3.decoder.key", "past_key_values.3.decoder.value", "past_key_values.3.encoder.key", "past_key_values.3.encoder.value", "past_key_values.4.decoder.key", "past_key_values.4.decoder.value", "past_key_values.4.encoder.key", "past_key_values.4.encoder.value", "past_key_values.5.decoder.key", "past_key_values.5.decoder.value", "past_key_values.5.encoder.key", "past_key_values.5.encoder.value", "past_key_values.6.decoder.key", "past_key_values.6.decoder.value", "past_key_values.6.encoder.key", "past_key_values.6.encoder.value", "past_key_values.7.decoder.key", "past_key_values.7.decoder.value", "past_key_values.7.encoder.key", "past_key_values.7.encoder.value", "past_key_values.8.decoder.key", "past_key_values.8.decoder.value", "past_key_values.8.encoder.key", "past_key_values.8.encoder.value", "past_key_values.9.decoder.key", "past_key_values.9.decoder.value", "past_key_values.9.encoder.key", "past_key_values.9.encoder.value", "past_key_values.10.decoder.key", "past_key_values.10.decoder.value", "past_key_values.10.encoder.key", "past_key_values.10.encoder.value", "past_key_values.11.decoder.key", "past_key_values.11.decoder.value", "past_key_values.11.encoder.key", "past_key_values.11.encoder.value"};
char* decPast_output_names[25] = {"logits", "present.0.decoder.key", "present.0.decoder.value", "present.1.decoder.key", "present.1.decoder.value", "present.2.decoder.key", "present.2.decoder.value", "present.3.decoder.key", "present.3.decoder.value", "present.4.decoder.key", "present.4.decoder.value", "present.5.decoder.key", "present.5.decoder.value", "present.6.decoder.key", "present.6.decoder.value", "present.7.decoder.key", "present.7.decoder.value", "present.8.decoder.key", "present.8.decoder.value", "present.9.decoder.key", "present.9.decoder.value", "present.10.decoder.key", "present.10.decoder.value", "present.11.decoder.key", "present.11.decoder.value"};
char* decPast_pkv_inputs[25] = {"", "past_key_values.0.decoder.key", "past_key_values.0.decoder.value", "past_key_values.1.decoder.key", "past_key_values.1.decoder.value", "past_key_values.2.decoder.key", "past_key_values.2.decoder.value", "past_key_values.3.decoder.key", "past_key_values.3.decoder.value", "past_key_values.4.decoder.key", "past_key_values.4.decoder.value", "past_key_values.5.decoder.key", "past_key_values.5.decoder.value", "past_key_values.6.decoder.key", "past_key_values.6.decoder.value", "past_key_values.7.decoder.key", "past_key_values.7.decoder.value", "past_key_values.8.decoder.key", "past_key_values.8.decoder.value", "past_key_values.9.decoder.key", "past_key_values.9.decoder.value", "past_key_values.10.decoder.key", "past_key_values.10.decoder.value", "past_key_values.11.decoder.key", "past_key_values.11.decoder.value"};

bool USE_GPU = false;
bool USING_F16_MODEL = true;   // true if using model with _Float16 values
//const char* CONFIG_PATH = "/home/tech/Documents/gitDir/GEC/gec-demo/src/native/gec_runtime/config/config.json";


int load_config(const char* configPath) {
    clock_t startTime = clock();

    Log(INFO, "Config Path: \x1b[1;92m%s\x1b[0m", configPath);
    // Read and parse the JSON file
    struct json_object *parsed_json;
    parsed_json = json_object_from_file(configPath);
    if (!parsed_json) {
        Log(ERROR, "Error parsing JSON file");
        return -1;
    }

    // Assign JSON values to variables
    struct json_object *path_onnx, *path_gibb, *log_level;
    char onnxPath[100];
    char gibbPath[100];

    if (json_object_object_get_ex(parsed_json, "logLevel", &log_level))
        LOG_LEVEL = json_object_get_int(log_level);

    if (json_object_object_get_ex(parsed_json, "gibberish_model_path", &path_gibb))
        strcpy(gibbPath, json_object_get_string(path_gibb));

    if (json_object_object_get_ex(parsed_json, "gec_model_path", &path_onnx)) 
        strcpy(onnxPath, json_object_get_string(path_onnx));
    
    snprintf(PATH_ENCODER, sizeof(PATH_ENCODER), "%s%s", onnxPath, "/encoder_model.onnx");
    snprintf(PATH_DECODER, sizeof(PATH_DECODER), "%s%s", onnxPath, "/decoder_model.onnx");
    snprintf(PATH_DECODER_PAST, sizeof(PATH_DECODER_PAST), "%s%s", onnxPath, "/decoder_with_past_model.onnx");
    snprintf(PATH_SP_MODEL, sizeof(PATH_SP_MODEL), "%s%s", onnxPath, "/spiece.model");
    snprintf(PATH_GIBB, sizeof(PATH_GIBB), "%s%s", gibbPath, "/model.onnx");
    snprintf(PATH_GIBB_TOKENIZER, sizeof(PATH_GIBB_TOKENIZER), "%s%s", gibbPath, "/tokenizer.json");

    // Close the file and free memory
    json_object_put(parsed_json);
    SetTimerValue("Load Config", startTime, clock());
    return 0;
}

// Initialize a new GECO instance
void* NewGeco(const char* configPath, int useGpu, int gpuId) {
    clock_t startTime = clock();
    Log(DEBUG, "Initializing a new Geco object...");

    // Load config
    if (load_config(configPath) != 0) {
        Log(ERROR, "Failed loading config!");
        return NULL;
    }

    // Allocate memory for GECO object
    Geco* geco = (Geco*)malloc(sizeof(Geco));
    if (geco == NULL) {
        Log(ERROR, "Failed to allocate memory for Geco object!");
        return NULL;
    }
    
    // Set all fields to 0 or NULL
    memset(geco, 0, sizeof(Geco));

    // Get the ONNX Runtime API handle
    geco->g_ort = OrtGetApiBase()->GetApi(ORT_API_VERSION);
    if (geco->g_ort == NULL) {
        Log(ERROR, "Failed to get ONNX Runtime API handle!");
        free(geco);
        geco = NULL;
        return NULL;
    }

    // Set device_id
    if (useGpu) USE_GPU = true;
    snprintf(geco->device_id, sizeof(geco->device_id), USE_GPU ? "gpu:%d" : "cpu", gpuId);
    Log(DEBUG, "  - Using Device: \"%s\"", geco->device_id);

    // Create an ONNX Runtime environment
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateEnv(ORT_LOGGING_LEVEL_WARNING, "test", &geco->env));
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateCpuMemoryInfo(OrtArenaAllocator, OrtMemTypeDefault, &geco->memory_info));
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateSessionOptions(&geco->session_options));  // Session options
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateRunOptions(&geco->run_options));          // Run options

    // Arena Configuration
    const char* keys[] = {"arena_extend_strategy"}; 
    const size_t values[] = {0};
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateArenaCfgV2(keys, values, 1, &geco->arena_cfg));

    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->SetIntraOpNumThreads(geco->session_options, 4));   // 1
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->SetInterOpNumThreads(geco->session_options, 2));
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->SetSessionGraphOptimizationLevel(geco->session_options, ORT_ENABLE_ALL));
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->DisableMemPattern(geco->session_options)); // Should prevent some fragmentation & Stops all logging messages "block in memory pattern size is: XXXX but the actual size is: XXXX"
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->EnableCpuMemArena(geco->session_options));

    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->AddSessionConfigEntry(geco->session_options, "session.use_env_allocators", "1"));       // Use the environment's allocators
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->AddSessionConfigEntry(geco->session_options, "session.dynamic_block_base", "4"));       // Improves gpu performance
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->AddSessionConfigEntry(geco->session_options, "session.use_device_allocator_for_initializers", "1"));

    if (USE_GPU) {
        // Create CUDA Memory info
        OrtCUDAProviderOptionsV2* cuda_options = NULL;
        ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateCUDAProviderOptions(&cuda_options));

        char idStr[5];
        snprintf(idStr, sizeof(idStr), "%d", gpuId);
        const char* provider_keys[] = {"device_id", "arena_extend_strategy"};
        const char* provider_values[] = {idStr, "kNextPowerOfTwo"};
        
        ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->UpdateCUDAProviderOptions(cuda_options, provider_keys, provider_values, 2));
        ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->SessionOptionsAppendExecutionProvider_CUDA_V2(geco->session_options, cuda_options));

        // Release after registering
        geco->g_ort->ReleaseCUDAProviderOptions(cuda_options);
    } else {
        // Create Memory Info for CPU
        ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->AddRunConfigEntry(geco->run_options, "memory.enable_memory_arena_shrinkage", "cpu:0"));
    }

    // Initialize Allocators & Sessions
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateSession(geco->env, PATH_ENCODER, geco->session_options, &geco->encoder_session));
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateSession(geco->env, PATH_DECODER, geco->session_options, &geco->decoder_session));
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateSession(geco->env, PATH_DECODER_PAST, geco->session_options, &geco->decPast_session));
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateSession(geco->env, PATH_GIBB, geco->session_options, &geco->gibb_session));

    // Sinlge shared allocator
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateAllocator(geco->encoder_session, geco->memory_info, &geco->allocator));

    // IO Bindings
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateIoBinding(geco->encoder_session, &geco->enc_io_binding));
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateIoBinding(geco->decoder_session, &geco->dec_io_binding));
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateIoBinding(geco->decPast_session, &geco->decPast_io_binding));
    ORT_CLEAN_ON_ERROR(init_fail, geco, geco->g_ort->CreateIoBinding(geco->gibb_session, &geco->gibb_io_binding));


    // Release session options
    geco->g_ort->ReleaseSessionOptions(geco->session_options);

    // Load the SentencePiece model
    geco->processor = initialize_processor(PATH_SP_MODEL);
    Log(DEBUG, "Geco has been created!");
    SetTimerValue("Create Geco", startTime, clock());
    return (void*)geco;

    // If the setup fails, free the allocated memory and return NULL
    init_fail:
    Log(ERROR, "Failed to create Geco object!");
    FreeGeco((void*)geco);
    return NULL;
}

// Frees all of the allocated memory used in this GECO object
void FreeGeco(void* objPtr) {
    clock_t startTime = clock();

    Log(INFO, "Freeing Geco object...");
    Geco* geco = (Geco*)objPtr;   // Cast the void* to Geco*
    if (geco == NULL) {
        Log(WARNING, "Got a NULL Geco! Nothing to free");
        return;
    }

    // Macro to help release resources
    #define RELEASE_RESOURCE(resource, release_func) \
        if (geco->resource) { \
            geco->g_ort->release_func(geco->resource); \
            geco->resource = NULL; \
        } else { \
            Log(WARNING, #resource " is NOT released"); \
        }



    // Check if each component is non-NULL before freeing/releasing
    if (geco->g_ort) {
        RELEASE_RESOURCE(input_tensor, ReleaseValue);
        RELEASE_RESOURCE(output_tensor, ReleaseValue);

        // Clear bounded input and outputs
        if (geco->enc_io_binding != NULL) {
            geco->g_ort->ClearBoundInputs(geco->enc_io_binding);
            geco->g_ort->ClearBoundOutputs(geco->enc_io_binding);
        }
        if (geco->dec_io_binding != NULL) {
            geco->g_ort->ClearBoundInputs(geco->dec_io_binding);
            geco->g_ort->ClearBoundOutputs(geco->dec_io_binding);
        }
        if (geco->decPast_io_binding != NULL) {
            geco->g_ort->ClearBoundInputs(geco->decPast_io_binding);
            geco->g_ort->ClearBoundOutputs(geco->decPast_io_binding);
        }
        if (geco->gibb_io_binding != NULL) {
            geco->g_ort->ClearBoundInputs(geco->gibb_io_binding);
            geco->g_ort->ClearBoundOutputs(geco->gibb_io_binding);
        }

        RELEASE_RESOURCE(enc_io_binding, ReleaseIoBinding);
        RELEASE_RESOURCE(dec_io_binding, ReleaseIoBinding);
        RELEASE_RESOURCE(decPast_io_binding, ReleaseIoBinding);
        RELEASE_RESOURCE(gibb_io_binding, ReleaseIoBinding);

        RELEASE_RESOURCE(allocator, ReleaseAllocator);

        RELEASE_RESOURCE(encoder_session, ReleaseSession);
        RELEASE_RESOURCE(decoder_session, ReleaseSession);
        RELEASE_RESOURCE(decPast_session, ReleaseSession);
        RELEASE_RESOURCE(gibb_session, ReleaseSession);
        
        RELEASE_RESOURCE(run_options, ReleaseRunOptions);
        //RELEASE_RESOURCE(session_options, ReleaseSessionOptions);
        
        if (geco->arena_cfg != NULL) {
            ORT_CLEAN_ON_ERROR(arena_label, geco, geco->g_ort->UnregisterAllocator(geco->env, geco->memory_info));
            arena_label:
            geco->g_ort->ReleaseArenaCfg(geco->arena_cfg);
            geco->arena_cfg = NULL;
        }
        RELEASE_RESOURCE(cuda_memory_info, ReleaseMemoryInfo);
        RELEASE_RESOURCE(memory_info, ReleaseMemoryInfo);
        RELEASE_RESOURCE(env, ReleaseEnv);
    } else {
        Log(WARNING, "g_ort is preventing ANYTHING from being released");
    }


    if (geco->processor != NULL) {
        free_processor(geco->processor);
    } else {
        Log(WARNING, "processor is NOT released");
    }

    // Free the Geco struct itself
    free(geco);
    geco = NULL;
    Log(INFO, "Geco has been freed");
    SetTimerValue("Destroy Geco", startTime, clock());
}


// Release OrtValue* tensors
void free_tensor(Geco* geco, OrtValue** tensor) {
    if (tensor != NULL && *tensor != NULL) {
        geco->g_ort->ReleaseValue(*tensor);
        *tensor = NULL;
    }
}

// Release the input tensor
void free_inpTensor(Geco* geco) {
    geco->g_ort->ReleaseValue(geco->input_tensor);
    geco->input_tensor = NULL;
}

void free_binded_tensors(Geco* geco) {
    if (geco->binded_tensors != NULL) {
        for (int i=0; i<(int)geco->binded_tensors_len; i++) {
            geco->g_ort->ReleaseValue(geco->binded_tensors[i]);
            geco->binded_tensors[i] = NULL;
        }
        geco->allocator->Free(geco->allocator, geco->binded_tensors);
        geco->binded_tensors = NULL;
    }
}

// Checks if last 6 tokens in an array are the same or alternating between two distinct values
bool checkRepeating(int* arr, int lastInd) {
    if (arr == NULL) {
        Log(ERROR, "checkRepeating() Array pointer is NULL");
        return false;
    }
    if (lastInd < 5) {
        return false;
    }

    if (arr[lastInd] == arr[lastInd-2] && arr[lastInd] == arr[lastInd-4] && arr[lastInd-1] == arr[lastInd-3] && arr[lastInd-1] == arr[lastInd-5]) {
        return true;
    }
    return false;
}

// Find the most probable tokens for each sequence
int getMaxTokens(Geco* geco, int64_t* newTokens, int batchSize, int* completed_sequences, int runNum) {
    clock_t startTime = clock();
    // Check for NULL pointers
    if (!geco || !geco->binded_tensors[0] || !newTokens || !completed_sequences) {
        Log(ERROR, "NULL pointer detected in input arguments");
        return -1;
    }

    if (USING_F16_MODEL) {
        // Get the logits data as a readable array
        _Float16* logitData = NULL;
        if (geco->g_ort->GetTensorMutableData(geco->binded_tensors[0], (void**)&logitData) != NULL) {
            Log(ERROR, "Failed to get logits data");
            return -1;
        }

        // Find the most likely next token for each sequence    
        for (int seqNum = 0; seqNum < batchSize; seqNum++) {
            if (completed_sequences[seqNum] == 1) {
                // This sequence is already completed, so add a 0 and continue to the next seq
                newTokens[seqNum] = 0;
                continue;
            }

            // If the sequence is starting a repeating loop then mark the indexes we won't allow to be generated next
            int ignore_indexes[2] = {-1, -1};
            if (checkRepeating(geco->generated_tokens[seqNum], runNum-1)) {
                ignore_indexes[0] = geco->generated_tokens[seqNum][runNum-1];
                ignore_indexes[1] = geco->generated_tokens[seqNum][runNum-2];
            }

            int start_index = seqNum * LOGIT_SIZE;
            _Float16 maxVal = logitData[start_index];
            int nextToken = 0;

            for (int i = (start_index+1); i < (start_index+LOGIT_SIZE); i++) {
                if (logitData[i] > maxVal) {
                    if ((i-start_index) == ignore_indexes[0] || (i-start_index) == ignore_indexes[1]) {
                        continue;   // Token is in the ignore list, so skip it
                    }
                    maxVal = logitData[i];
                    nextToken = i - start_index;
                }
            }
            // Add to array of generated tokens
            newTokens[seqNum] = (int64_t)nextToken;
            
            // If this sequence reaches its eos token, mark it as completed
            if (nextToken == 1) {
                completed_sequences[seqNum] = 1;
            }
        }
        logitData = NULL;
    } else {
        // Get the logits data as a readable array
        float* logitData = NULL;
        if (geco->g_ort->GetTensorMutableData(geco->binded_tensors[0], (void**)&logitData) != NULL) {
            Log(ERROR, "Failed to get logits data");
            return -1;
        }

        // Find the most likely next token for each sequence    
        for (int seqNum = 0; seqNum < batchSize; seqNum++) {
            if (completed_sequences[seqNum] == 1) {
                // This sequence is already completed, so add a 0 and continue to the next seq
                newTokens[seqNum] = 0;
                continue;
            }

            // If the sequence is starting a repeating loop then mark the indexes we won't allow to be generated next
            int ignore_indexes[2] = {-1, -1};
            if (checkRepeating(geco->generated_tokens[seqNum], runNum-1)) {
                ignore_indexes[0] = geco->generated_tokens[seqNum][runNum-1];
                ignore_indexes[1] = geco->generated_tokens[seqNum][runNum-2];
            }

            int start_index = seqNum * LOGIT_SIZE;
            float maxVal = logitData[start_index];
            int nextToken = 0;

            for (int i = (start_index+1); i < (start_index+LOGIT_SIZE); i++) {
                if (logitData[i] > maxVal) {
                    if ((i-start_index) == ignore_indexes[0] || (i-start_index) == ignore_indexes[1]) {
                        continue;   // Token is in the ignore list, so skip it
                    }
                    maxVal = logitData[i];
                    nextToken = i - start_index;
                }
            }
            // Add to array of generated tokens
            newTokens[seqNum] = (int64_t)nextToken;
            
            // If this sequence reaches its eos token, mark it as completed
            if (nextToken == 1) {
                completed_sequences[seqNum] = 1;
            }
        }
        logitData = NULL;
    }

    SetTimerValue("Get Max Tokens", startTime, clock());
    return 0;
}


// Recursively run the Decoder with past model
void runPast(Geco* geco, int runNum, int64_t* nextToks, int batchSize, int* completed_sequences) {
    // Run the Model with IO Bindings and get the output tensors
    struct timespec gpuTime;
    clock_gettime(CLOCK_MONOTONIC, &gpuTime);
    ORT_CLEAN_ON_ERROR(decPast_cleanup, geco, geco->g_ort->RunWithBinding(geco->decPast_session, geco->run_options, geco->decPast_io_binding));
    SetGpuTime(gpuTime);
    ORT_CLEAN_ON_ERROR(decPast_cleanup, geco, geco->g_ort->GetBoundOutputValues(geco->decPast_io_binding, geco->allocator, &geco->binded_tensors, &geco->binded_tensors_len));

    // Get the most likely next tokens
    if (getMaxTokens(geco, nextToks, batchSize, completed_sequences, runNum)) {
        goto decPast_cleanup;
    }

    // Add the new tokens to the generated tokens array
    bool sequences_completed = true;
    for (int i = 0; i < batchSize; i++) {
        geco->generated_tokens[i][runNum] = nextToks[i];
        if (completed_sequences[i] != 1) {
            sequences_completed = false;
        }
    }
  
    // Check if all sequences are finished
    if (sequences_completed || (runNum+1) == MAX_TOKENS) { 
        //Log(DEBUG, "All sequences completed @ Run #%d!\n", runNum); 
        goto decPast_cleanup;
    }

    // Bind the input tensor: "input_ids"
    ORT_CLEAN_ON_ERROR(decPast_cleanup, geco, geco->g_ort->CreateTensorWithDataAsOrtValue(geco->memory_info, nextToks, (batchSize*sizeof(int64_t)), (int64_t[]){batchSize, 1}, 2, ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, &geco->input_tensor));
    ORT_CLEAN_ON_ERROR(decPast_cleanup, geco, geco->g_ort->BindInput(geco->decPast_io_binding, "input_ids", geco->input_tensor));
    free_inpTensor(geco);

    // Clear bindings before setting new ones
    geco->g_ort->ClearBoundOutputs(geco->decPast_io_binding);

    // Bind the output tensors to the input tensors & Create new shaped output tensors
    for (int i=0; i<25; i++) {
        // Updates the output tensors shapes
        ORT_CLEAN_ON_ERROR(decPast_cleanup, geco, geco->g_ort->BindOutputToDevice(geco->decPast_io_binding, decPast_output_names[i], geco->memory_info));

        // Update the input tensor bindings with the previous output tensors (Skip 'logits' since it has no correlating input)
        if (i != 0) {
            // Changes output_name from "present.0.decoder.key" to "past_key_values.0.decoder.key"
            ORT_CLEAN_ON_ERROR(decPast_cleanup, geco, geco->g_ort->BindInput(geco->decPast_io_binding, decPast_pkv_inputs[i], geco->binded_tensors[i]));
        }
        geco->g_ort->ReleaseValue(geco->binded_tensors[i]);
        geco->binded_tensors[i] = NULL;
    }
    geco->allocator->Free(geco->allocator, geco->binded_tensors);
    geco->binded_tensors = NULL;

    // RECURSE
    runPast(geco, runNum+1, nextToks, batchSize, completed_sequences);
    
    // Free memory
    decPast_cleanup:
    free_binded_tensors(geco);
    free_inpTensor(geco);
    geco->g_ort->ClearBoundInputs(geco->decPast_io_binding);
    geco->g_ort->ClearBoundOutputs(geco->decPast_io_binding);
}

// Run the both decoder models
void runDecoders(Geco* geco, int batchSize) {
    clock_t startTime = clock();
    geco->binded_tensors = NULL;
    geco->binded_tensors_len = 0;
    
    // Token arrays
    int64_t newTokens[MAX_BATCH_SIZE];                  // Array to hold the newly generated tokens
    int completed_sequences[MAX_BATCH_SIZE] = {0};      // Mark 1 if a sequence is completed (Otherwise they all are set to 0 values)
    int64_t init_tokens[MAX_BATCH_SIZE] = {0};          // Initial tokens of 0's (Must be calloc to be the correct size for the tensor creation)

    // Decoder "input_ids" variables
    size_t inputs_data_len = batchSize * sizeof(int64_t);
    int64_t inputs_shape[2] = {batchSize, 1};
    ORT_CLEAN_ON_ERROR(decoder_cleanup, geco, geco->g_ort->CreateTensorWithDataAsOrtValue(geco->memory_info, init_tokens, inputs_data_len, inputs_shape, 2, ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, &geco->input_tensor));
    ORT_CLEAN_ON_ERROR(decoder_cleanup, geco, geco->g_ort->BindInput(geco->dec_io_binding, "input_ids", geco->input_tensor));
    free_inpTensor(geco);
    
    // Bind the Output Tensors
    for (int i = 0; i < 49; i++) {
        ORT_CLEAN_ON_ERROR(decoder_cleanup, geco, geco->g_ort->BindOutputToDevice(geco->dec_io_binding, decoder_output_names[i], geco->memory_info));
    }

    // Run the Model
    struct timespec gpuTime;
    clock_gettime(CLOCK_MONOTONIC, &gpuTime);
    ORT_CLEAN_ON_ERROR(decoder_cleanup, geco, geco->g_ort->RunWithBinding(geco->decoder_session, geco->run_options, geco->dec_io_binding));
    SetGpuTime(gpuTime);

    // Get the return output values as OrtValue* Tensors
    ORT_CLEAN_ON_ERROR(decoder_cleanup, geco, geco->g_ort->GetBoundOutputValues(geco->dec_io_binding, geco->allocator, &geco->binded_tensors, &geco->binded_tensors_len));
    if (getMaxTokens(geco, newTokens, batchSize, completed_sequences, 1)) {
        goto decoder_cleanup;
    }

    // Add the new tokens to the generated tokens array
    for (int i = 0; i < batchSize; i++) {
        geco->generated_tokens[i][1] = newTokens[i];
    }
    SetTimerValue("Main Decoder", startTime, clock());


    // DECODER_WITH_PAST_MODEL
    // Bind the input tensors to IO bindings
    ORT_CLEAN_ON_ERROR(decoder_cleanup, geco, geco->g_ort->CreateTensorWithDataAsOrtValue(geco->memory_info, newTokens, inputs_data_len, inputs_shape, 2, ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, &geco->input_tensor));
    ORT_CLEAN_ON_ERROR(decoder_cleanup, geco, geco->g_ort->BindInput(geco->decPast_io_binding, "input_ids", geco->input_tensor));
    free_inpTensor(geco);
    for (int i = 3; i < 51; i++) {
        // Append the Decoders outputs as inputs to the DecPast model
        ORT_CLEAN_ON_ERROR(decoder_cleanup, geco, geco->g_ort->BindInput(geco->decPast_io_binding, decPast_input_names[i], geco->binded_tensors[i-2]));
    }

    // RELEASE ALL BINDED TENSORS
    free_binded_tensors(geco);
    geco->g_ort->ClearBoundOutputs(geco->dec_io_binding);

    // Create & Bind the Output Tensors
    for (int i = 0; i < 25; i++) {
        ORT_CLEAN_ON_ERROR(decoder_cleanup, geco, geco->g_ort->BindOutputToDevice(geco->decPast_io_binding, decPast_output_names[i], geco->memory_info));
    }

    // Run and recurse
    clock_t rp_startTime = clock();
    runPast(geco, 2, newTokens, batchSize, completed_sequences);
    SetTimerValue("Past Decoder", rp_startTime, clock());

    // Clean up
    decoder_cleanup:
    free_binded_tensors(geco);
    free_inpTensor(geco);
    geco->g_ort->ClearBoundInputs(geco->dec_io_binding);
    geco->g_ort->ClearBoundOutputs(geco->dec_io_binding);
    SetTimerValue("All Decoders", startTime, clock());
}

// Uses the float tensors data (`output_tensor`) to create a new tensor of _Float16 data
int extract_float16_from_tensor(Geco* geco) {
    float* float_data;
    _Float16* float16_data;

    // Get tensor type and shape
    OrtTensorTypeAndShapeInfo* shape_info;
    ORT_CLEAN_ON_ERROR(extract_error_clean, geco, geco->g_ort->GetTensorTypeAndShape(geco->output_tensor, &shape_info));

    // Get dimensions and shape array
    size_t num_dims; // 3 Dimensions
    ORT_CLEAN_ON_ERROR(extract_error_clean, geco, geco->g_ort->GetDimensionsCount(shape_info, &num_dims));

    int64_t* tensor_shape = (int64_t*)malloc(num_dims * sizeof(int64_t));
    ORT_CLEAN_ON_ERROR(extract_error_clean, geco, geco->g_ort->GetDimensions(shape_info, tensor_shape, num_dims));

    size_t total_len;
    ORT_CLEAN_ON_ERROR(extract_error_clean, geco, geco->g_ort->GetTensorShapeElementCount(shape_info, &total_len));

    // Get pointer to float data
    ORT_CLEAN_ON_ERROR(extract_error_clean, geco, geco->g_ort->GetTensorMutableData(geco->output_tensor, (void**)&float_data));

    // Allocate memory for _Float16 array
    float16_data = (_Float16*)malloc(sizeof(_Float16) * total_len);
    if (!float16_data) {
        fprintf(stderr, "Memory allocation for float16 failed");
        goto extract_error_clean;
    }

    // Convert float32 to float16
    for (size_t i = 0; i < total_len; i++) {
        float16_data[i] = (_Float16)float_data[i];
    }

    // _Float16 is 2 bytes per element
    size_t total_len_fp16 = total_len*2;
    ORT_CLEAN_ON_ERROR(extract_error_clean, geco, geco->g_ort->CreateTensorWithDataAsOrtValue(geco->memory_info, float16_data, total_len_fp16, tensor_shape, 3, ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT16, &geco->output_tensor_fp16));

    geco->g_ort->ReleaseTensorTypeAndShapeInfo(shape_info);
    return 0;

    extract_error_clean:
    geco->g_ort->ReleaseTensorTypeAndShapeInfo(shape_info);
    return -1;
}

// Function to execute inference using the Geco context
void GecoRun(void* context, char** texts, int num_texts, char** result) {
    if (context == NULL) {
        Log(ERROR, "Invalid Geco context!");
        return;
    }

    // Cast the context to Geco pointer
    Geco* geco = (Geco*)context;

    // Call the InferModel function with the geco and input texts
    Log(DEBUG, "Infer GEC on device '%s'", geco->device_id);
    InferModel(geco, texts, num_texts, result);
}

// Implementation of the InferModel function
void InferModel(Geco* geco, char** texts, int num_texts, char** result) {
    clock_t startTime = clock();
    geco->input_tensor = NULL;
    geco->output_tensor = NULL;
    geco->output_tensor_fp16 = NULL;

    // Group and tokenize the texts
    TokenizedTexts *tokTexts = prepare_texts(geco->processor, texts, num_texts);
    if (tokTexts == NULL) {
        Log(ERROR, "Failed to create the TokenizedTexts object");
        goto infer_cleanup;
    }
    SetTimerValue("GEC Preproc", startTime, clock());

    // If their are no tokens to process, return NULL
    if (tokTexts->shape[0]*tokTexts->shape[1] == 0) {
        Log(ERROR, "No tokens to process, returning an empty string");
        goto infer_cleanup;
    }

    int batchSize = (int)tokTexts->shape[0];
    if (batchSize > MAX_BATCH_SIZE) {
        Log(ERROR, "batch size is too large to process: %d > %d", batchSize, MAX_BATCH_SIZE);
        goto infer_cleanup;
    }

    // Reset the generated tokens array to all 0's
    for (int i = 0; i < MAX_BATCH_SIZE; i++) {
        for (int j = 0; j < MAX_TOKENS; j++) {
            geco->generated_tokens[i][j] = 0;
        }
    }


    // Create & Bind the Input/Output tensors
    // Tensor: "attention_mask"
    ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->CreateTensorWithDataAsOrtValue(geco->memory_info, tokTexts->attention_mask, tokTexts->data_len, tokTexts->shape, 2, ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, &geco->input_tensor));
    ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->BindInput(geco->enc_io_binding, "attention_mask", geco->input_tensor));
    ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->BindInput(geco->dec_io_binding, "encoder_attention_mask", geco->input_tensor));
    ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->BindInput(geco->decPast_io_binding, "encoder_attention_mask", geco->input_tensor));
    free_inpTensor(geco);

    // Tensor: "input_ids"
    ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->CreateTensorWithDataAsOrtValue(geco->memory_info, tokTexts->ids, tokTexts->data_len, tokTexts->shape, 2, 7, &geco->input_tensor));
    ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->BindInput(geco->enc_io_binding, "input_ids", geco->input_tensor));
    free_inpTensor(geco);

    // Tensor: "last_hidden_state"
    int64_t output_shape[3] = {tokTexts->shape[0], tokTexts->shape[1], 768};
    ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->CreateTensorAsOrtValue(geco->allocator, output_shape, 3, ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT, &geco->output_tensor));
    ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->BindOutput(geco->enc_io_binding, "last_hidden_state", geco->output_tensor));

    // Run the Encoder
    struct timespec gpuTime;
    clock_gettime(CLOCK_MONOTONIC, &gpuTime);
    ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->RunWithBinding(geco->encoder_session, geco->run_options, geco->enc_io_binding));
    SetGpuTime(gpuTime);
    
    if (USING_F16_MODEL) {
        // Create new tensor with float16 data from 'last_hidden_state' tensor
        Log(DEBUG, "Creating new tensor 'output_tensor_fp16'");
        if (extract_float16_from_tensor(geco) == -1) {
            fprintf(stderr, "Tensor conversion to _Float16 failed");
            goto infer_cleanup;
        }

        // Apply new tensor to the decoders, then free it
        Log(DEBUG, "Binding 'encoder_hidden_states' for F16 model");
        ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->BindInput(geco->dec_io_binding, "encoder_hidden_states", geco->output_tensor_fp16));
        Log(DEBUG, "Bound 'encoder_hidden_states' to decoder input for F16 model!");
    } else {
        // Using original model type with Float32 values
        Log(DEBUG, "Binding 'encoder_hidden_states' for F32 model");
        ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->BindInput(geco->dec_io_binding, "encoder_hidden_states", geco->output_tensor));
        ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->BindInput(geco->decPast_io_binding, "encoder_hidden_states", geco->output_tensor));
        Log(DEBUG, "Bound 'encoder_hidden_states' to decoder input for F32 model!");
    }
    //ORT_CLEAN_ON_ERROR(infer_cleanup, geco, geco->g_ort->BindInput(geco->decPast_io_binding, "encoder_hidden_states", geco->output_tensor));
    geco->g_ort->ReleaseValue(geco->output_tensor);
    geco->output_tensor = NULL;

    // Run Decoder and Decoder-With-Past model sessions
    runDecoders(geco, batchSize);
    Log(DEBUG, "Decoders finished running");


    // Decode and print the results
    clock_t postTime = clock();
    *result = decode_texts(geco->processor, geco->generated_tokens, tokTexts);
    SetTimerValue("GEC Postproc", postTime, clock());
    
    // CLEAN UP
    infer_cleanup:
    free_tokenized_texts(tokTexts);
    free_binded_tensors(geco);
    free_inpTensor(geco);
    geco->g_ort->ReleaseValue(geco->output_tensor);
    geco->output_tensor = NULL;
    geco->g_ort->ReleaseValue(geco->output_tensor_fp16);
    geco->output_tensor_fp16 = NULL;
    geco->g_ort->ClearBoundInputs(geco->enc_io_binding);
    geco->g_ort->ClearBoundOutputs(geco->enc_io_binding);
    geco->g_ort->ClearBoundInputs(geco->dec_io_binding);
    geco->g_ort->ClearBoundOutputs(geco->dec_io_binding);
    geco->g_ort->ClearBoundInputs(geco->decPast_io_binding);
    geco->g_ort->ClearBoundOutputs(geco->decPast_io_binding);
    SetTimerValue("GEC Total", startTime, clock());
    //PrintTimes();
}


// Function to execute inference using the Geco context
void GecoGibb(void* context, double probs[MAX_BATCH_SIZE][GIBB_CLASSES], char** texts, int num_batches) {
    // Check if the context is valid
    if (context == NULL) {
        Log(ERROR, "Invalid Gibberish context!");
        return;
    }

    // Cast the context to Geco pointer
    Geco* geco = (Geco*)context;

    // Call the InferGibb function with the geco and input texts
    InferGibb(geco, probs, texts, num_batches);
}

void InferGibb(Geco* geco, double probs[MAX_BATCH_SIZE][GIBB_CLASSES], char** texts, int num_batches) {
    clock_t startTime = clock();
    geco->input_tensor = NULL;
    geco->output_tensor = NULL;
    Tokenized_WP_Output tokenized_texts;

    // Tokenize the input texts
    tokenized_texts = batch_gibb_texts(PATH_GIBB_TOKENIZER, texts, num_batches);
    SetTimerValue("Gibb Preproc", startTime, clock());
    int batchSize = tokenized_texts.shape[0];

    // Reset probs array to 0
    for (int i=0; i<MAX_BATCH_SIZE; i++) {
        for (int j=0; j<GIBB_CLASSES; j++) {
            probs[i][j] = 0.0;
        }
    }

    // Tensor: "input_ids"
    int64_t data_len = tokenized_texts.shape[0] * tokenized_texts.shape[1] * sizeof(int64_t);
    ORT_CLEAN_ON_ERROR(gibberish_cleanup, geco, geco->g_ort->CreateTensorWithDataAsOrtValue(
        geco->memory_info, 
        tokenized_texts.ids, 
        data_len, 
        tokenized_texts.shape, 
        2, 
        ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, 
        &geco->input_tensor
    ));
    ORT_CLEAN_ON_ERROR(gibberish_cleanup, geco, geco->g_ort->BindInput(geco->gibb_io_binding, "input_ids", geco->input_tensor));
    free_inpTensor(geco);

    // Tensor: "attention_mask"
    ORT_CLEAN_ON_ERROR(gibberish_cleanup, geco, geco->g_ort->CreateTensorWithDataAsOrtValue(
        geco->memory_info, 
        tokenized_texts.attention_mask, 
        data_len, 
        tokenized_texts.shape, 
        2, 
        ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, 
        &geco->input_tensor
    ));
    ORT_CLEAN_ON_ERROR(gibberish_cleanup, geco, geco->g_ort->BindInput(geco->gibb_io_binding, "attention_mask", geco->input_tensor));
    free_inpTensor(geco);


    // Tensor: "logits"
    int64_t logit_shape[2] = {batchSize, GIBB_CLASSES};
    ORT_CLEAN_ON_ERROR(gibberish_cleanup, geco, geco->g_ort->CreateTensorAsOrtValue(geco->allocator, logit_shape, 2, ONNX_TENSOR_ELEMENT_DATA_TYPE_FLOAT, &geco->output_tensor));
    ORT_CLEAN_ON_ERROR(gibberish_cleanup, geco, geco->g_ort->BindOutput(geco->gibb_io_binding, "logits", geco->output_tensor));

    // Run the model
    clock_t gibbTime = clock();
    struct timespec gpuTime;
    clock_gettime(CLOCK_MONOTONIC, &gpuTime);
    ORT_CLEAN_ON_ERROR(gibberish_cleanup, geco, geco->g_ort->RunWithBinding(geco->gibb_session, geco->run_options, geco->gibb_io_binding));
    SetGpuTime(gpuTime);
    SetTimerValue("Gibb Model Run", gibbTime, clock());

    // Perform softmax on an array of logits to convert them to probabilities
    clock_t softmaxTime = clock();
    float* logitData;
    ORT_CLEAN_ON_ERROR(gibberish_cleanup, geco, geco->g_ort->GetTensorMutableData(geco->output_tensor, (void**)&logitData));

    for (int seqNum = 0; seqNum < batchSize; seqNum++) {
        int idx = seqNum * GIBB_CLASSES;
        int64_t max_logit = logitData[idx];

        // Find the maximum logit for numerical stability
        for (int i = 1; i < GIBB_CLASSES; i++) {
            if (logitData[idx+i] > max_logit) {
                max_logit = logitData[idx+i];
            }
        }

        // Compute exponentials and sum them
        double sum = 0.0;
        for (int i = 0; i < GIBB_CLASSES; i++) {
            probs[seqNum][i] = exp(logitData[idx+i] - max_logit); // Subtract max_logit for numerical stability
            sum += probs[seqNum][i];
        }

        // Normalize the probabilities
        for (int i = 0; i < GIBB_CLASSES; i++) {
            probs[seqNum][i] /= sum;
            probs[seqNum][i] *= 100;
        }
    }
    SetTimerValue("Calculate Softmax", softmaxTime, clock());

    gibberish_cleanup:
    free_tokenized_wp_output(tokenized_texts);
    free_inpTensor(geco);
    if (geco->output_tensor != NULL) {
        geco->g_ort->ReleaseValue(geco->output_tensor);
        geco->output_tensor = NULL;
    }
    logitData = NULL;
    // Clear bindings
    geco->g_ort->ClearBoundInputs(geco->gibb_io_binding);
    geco->g_ort->ClearBoundOutputs(geco->gibb_io_binding);
    SetTimerValue("Gibberish Total", startTime, clock());
    malloc_trim(0); // Release unused data in the heap
    //printTimes();
    //Log(INFO, "Total Time in %s = %.3f", geco->device_id, allTimes[15]);
}

 