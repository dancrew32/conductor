package phase

import (
	"strconv"

	"github.com/Nextdoor/conductor/services/build"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/shared/settings"
	"github.com/Nextdoor/conductor/shared/types"
)

type jenkinsPhase struct{}

func newJenkins() *jenkinsPhase {
	return &jenkinsPhase{}
}

func (p *jenkinsPhase) Start(phase *types.Phase, buildUser *types.User) error {
	params := make(map[string]string)
	params["TRAIN_ID"] = strconv.FormatUint(phase.Train.ID, 10)
	params["BRANCH"] = phase.Train.Branch
	params["SHA"] = phase.PhaseGroup.HeadSHA
	params["CONDUCTOR_HOSTNAME"] = settings.GetHostname()
	if buildUser != nil {
		params["BUILD_USER"] = buildUser.Name
	} else {
		params["BUILD_USER"] = "Conductor"
	}

	var workflow Workflow
	switch phase.Type {
	case types.Delivery:
		workflow = deliveryWorkflow
	case types.Verification:
		workflow = verificationWorkflow
	case types.Deploy:
		workflow = deployWorkflow
	}

	if len(workflow.Stages) == 0 {
		return nil
	}

	// Trigger stage 1.
	stage := workflow.Stages[0]
	if len(stage.Jobs) == 0 {
		return nil
	}

	for _, jobName := range stage.Jobs {
		var jobToTrigger *types.Job = nil
		for _, job := range phase.Jobs {
			if job.Name == jobName {
				jobToTrigger = job
				break
			}
		}
		if jobToTrigger == nil {
			data.Client()
		}

		go build.Jenkins().TriggerJob(jobToTrigger, params)
	}

	return nil
}

// There are two different polling operations - One is polling known jobs for the currently active phase
// (also other running phases? How do we know all the currently running phases?), the other is polling for new jobs.
// What if we just poll all the currently running jobs from the jobs defined in the workflows + the jobs defined in the expected job list.
// Do we need to do workflows + expected jobs?
// Can we just use expected jobs?
// The UI is based on expected jobs... phase transitions are based on expected jobs... but it tracks and triggers the workflows separately.
