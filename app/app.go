package app

import (
	"fmt"
	"k8s-token-hunter/fileops"
	"k8s-token-hunter/internal"
	"k8s-token-hunter/model"
	"k8s-token-hunter/presenter"
	"log"
	"os"
	"time"

	"github.com/common-nighthawk/go-figure"
)

func Run(config model.Config) {
	myFigure := figure.NewColorFigure("K8S TOKEN HUNTER", "", "cyan", true)
	myFigure.Print()

	if internal.CheckPrerequisites() {
		fmt.Println("[+] All prerequisites are met.")
	} else {
		fmt.Println("[-] Some prerequisites are not met. Exiting.")
		os.Exit(1)
	}

	// Proceed with the rest of the program using flags...
	fmt.Printf("Parsed flags: %+v\n", config)

	startTime := time.Now()
	log.Printf("Start Time: %v\n", startTime.Format(time.RFC822))

	if config.ServiceAccountToken == "" {
		config.ServiceAccountToken = fileops.GetFileContent(internal.SaTokenPath)
	}

	if config.SuffixGeneratorFile != "" {
		log.Printf("Generating suffix list %v\n", config.SuffixGeneratorFile)
		err := internal.GenerateCombinationsToFile(config.SuffixGeneratorFile, 5)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during command execution: %v\n", err)
			os.Exit(1)
		}
	}

	results := internal.ExecuteBruteForce(&config)

	if len(results) > 0 {
		fileops.SaveResultToFile(results, &config.OutputPath)
		for _, result := range results {
			resultPresenter := presenter.ResultPresenter{Result: result}
			presenter.CreateAndRenderTable(resultPresenter)
		}
	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)

	log.Printf("Start Time: %v\n", startTime.Format(time.RFC822))
	log.Printf("End Time: %v\n", endTime.Format(time.RFC822))
	log.Printf("Elapsed Time: %v\n", elapsedTime.Seconds())
}
