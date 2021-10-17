package controllers

import (
	"fmt"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
	"strings"
)

type GolandReporter struct{}

func (g GolandReporter) SpecSuiteWillBegin(config.GinkgoConfigType, *types.SuiteSummary) {

}

func (g GolandReporter) BeforeSuiteDidRun(*types.SetupSummary) {

}

func (g GolandReporter) SpecWillRun(specSummary *types.SpecSummary) {
	fmt.Printf("=== RUN   %s\n", strings.Join(specSummary.ComponentTexts[1:], " "))
}

func (g GolandReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	if specSummary.Passed() {
		printResultOutput(specSummary, "PASS")
	} else if specSummary.HasFailureState() {
		fmt.Printf("%s\n\n", specSummary.Failure.Message)
		fmt.Printf("%s\n\n", specSummary.Failure.Location.FullStackTrace)
		printResultOutput(specSummary, "FAIL")
	} else if specSummary.Skipped() {
		printResultOutput(specSummary, "SKIP")
	} else if specSummary.Pending() {
		printResultOutput(specSummary, "SKIP")
	} else {
		panic("Unknown test output")
	}
}

func (g GolandReporter) AfterSuiteDidRun(*types.SetupSummary) {

}

func (g GolandReporter) SpecSuiteDidEnd(*types.SuiteSummary) {

}

func printResultOutput(specSummary *types.SpecSummary, result string) {
	fmt.Printf("--- %s: %s (%.3fs)\n", result, strings.Join(specSummary.ComponentTexts[1:], " "), specSummary.RunTime.Seconds())
}