package testprocessor

import (
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"github.com/distributed-containers-inc/knoci/pkg/controller/testrunner"
)

type StateRunning struct{}

func (s *StateRunning) Process(processor *TestProcessor) error {
	test, err := processor.getTest()
	if err != nil {
		return fmt.Errorf("could not get test: %s", err.Error())
	}

	runner := testrunner.New(
		processor.KubeCli,
		processor.TestNamespace,
		processor.TestName,
		processor.TestSpec,
		processor.ctx,
	)

	if test.Status == nil {
		return fmt.Errorf("test status was nil")
	}
	if test.Status.NumberOfTests <= 0 {
		runner.Parallelize = false
	} else {
		runner.NumberOfTests = test.Status.NumberOfTests
		runner.Parallelize = true
	}

	runner.SplittingTime = 10 //wait 10 seconds before killing & splitting

	err = runner.Run()
	if err != nil {
		return processor.setState(
			v1alpha1.StateFailed,
			fmt.Sprintf(
				"Error while running: %s",
				err.Error(),
			),
		)
	}
	return nil
}
