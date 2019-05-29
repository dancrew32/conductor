/*
 Workflow mechanism:


[
	[[build], [a, b, c]]
	[[tests]]
]

[
	[[deploy], [us1, eu1]]
]

{
	"delivery": "(build -> delivery) | tests",
	"verification": "(build -> delivery) | tests",
	"deploy": "(build -> delivery) | tests",
}

Build ID high water mark per job

((build -> delivery) | tests) -> integration

prod-deploy -> (us1 | eu1 | au1)

 DELIVERY_WORKFLOW=(build, delivery) + tests
 VERIFICATION_WORKFLOW=
 DEPLOY_WORKFLOW=deploy, deploy-us1 + deploy-eu1 + deploy-tag
 ROLLBACK_WORKFLOW=rollback,  deploy-us1 + deploy-eu1
 */

package phase

import (
	"strings"

	"github.com/Nextdoor/conductor/shared/flags"
)

var (
	// These are the names of the jobs which will be triggered in each phase.
	// Comma separates groups of jobs to run. The next group will only be run if everything prior to it finishes.
	// Plus means run these jobs in parallel.
	// For example: "a + b, c" means trigger a and b in parallel, wait for them to complete, and then run c.
	// A workflow for one phase can trigger jobs for another phase. The jobs linked to a phase are defined in the
	// DELIVERY_JOBS, VERIFICATION_JOBS, DEPLOY_JOBS, and ROLLBACK_JOBS variables.

	// Conductor will trigger the jobs, and watches their execution.
	// As the job runs, it will poll its progress and check its output once it finishes.
	// When a job completes, Conductor records that in its database.

	// How does Conductor resume polling a job after restart?
	// How does Conductor handle manual rebuilds or retriggers outside of its workflow control?
	// Do we need a retry or rebuild button on Conductor? - No, not for MVP.
	// Do we need some indicator of the stages on the Conductor UI itself? - No, not for MVP.
	// How does Conductor discover that a job finished while it was restarting? Does it remember the build id it was polling?
	// Currently the model stores the SHA of the phase group. It could try to find the last job with the phase group that matched...
	// Or we can store the build id in the database. Then it can query it from Jenkins and see if it can find it.
	// If it cannot find it, will it automatically restart? Or is that a manual action?
	deliveryWorkflowRaw     = flags.EnvString("DELIVERY_WORKFLOW", "")
	verificationWorkflowRaw = flags.EnvString("VERIFICATION_WORKFLOW", "")
	deployWorkflowRaw       = flags.EnvString("DEPLOY_WORKFLOW", "")
	rollbackWorkflowRaw       = flags.EnvString("ROLLBACK_WORKFLOW", "")

	deliveryWorkflow Workflow
	verificationWorkflow Workflow
	deployWorkflow Workflow
	rollbackWorkflow Workflow
)

type Workflow struct {
	Stages []Stage
}

type Stage struct {
	Jobs []string
}

func parseWorkflows() {
	deliveryWorkflow = parseWorkflow(deliveryWorkflowRaw)
	verificationWorkflow = parseWorkflow(deliveryWorkflowRaw)
	deployWorkflow = parseWorkflow(deliveryWorkflowRaw)
	rollbackWorkflow = parseWorkflow(deliveryWorkflowRaw)
}

func parseWorkflow(workflowRaw string) Workflow {
	stages := make([]Stage, 0)

	workflowStages := strings.Split(workflowRaw, ",")
	for _, workflowStage := range workflowStages {
		jobs := make([]string, 0)

		workflowStage = strings.TrimSpace(workflowStage)
		stageJobs := strings.Split(workflowStage, "+")
		for _, stageJob := range stageJobs {
			stageJob = strings.TrimSpace(stageJob)
			jobs = append(jobs, stageJob)
		}

		stages = append(stages, Stage{Jobs: jobs})
	}

	return Workflow{Stages: stages}
}
