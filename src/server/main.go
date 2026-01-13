package main

import (
	"fmt"
	//	"speechtagger"
	"gec-api/speechtagger"
	"gec-api/print"
)

func main() {
	//StartServer()
	test_printer()
}

func test_printer() {
	// Default level is INFO, so only INFO and above will print
	fmt.Printf("Log-Level: %v\n", print.GetLevel())
	print.SetLevel(0)
	fmt.Printf("Log-Level: %v\n", print.GetLevel())
	print.Debug("This debug message won't show")
	print.Info("This info message will show")
	print.Warning("This warning will show")
	print.Error("This error will show in red")
	
	// Change log level to DEBUG
	print.SetLevel(print.LevelDebug)
	print.Debug("Now debug messages will show!")
	
	// Or use helper functions
	print.SetLevelWarning()
	print.Info("This won't show at WARNING level")
	print.Warning("But this will!")
	
	// Change to ERROR level only
	print.SetLevelError()
	print.Warning("This warning won't show")
	print.Error("Only errors show now")

}

func test_speechtagger() {
	// Call the InitTaggingModel function
	err := speechtagger.InitTaggingModel()

	// Check if the initialization returned an error
	if err != nil {
		fmt.Printf("InitTaggingModel() returned an error: %v\n", err)
	}

	// Check if the TaggerModel is initialized
	if speechtagger.TaggerModel == nil {
		fmt.Printf("TaggerModel is nil after initialization\n")
	}

	// Check if the SentTokenizer is initialized
	if speechtagger.SentTokenizer == nil {
		fmt.Printf("SentTokenizer is nil after initialization\n")
	}

	if speechtagger.TagsGob == "" || speechtagger.WeightsGob == "" {
		fmt.Printf("ERROR: TagsGob or WeightsGob is empty\n")
	}
	fmt.Printf("All passed!\n")
}
