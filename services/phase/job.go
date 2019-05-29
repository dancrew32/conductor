package phase

import (
	"sort"

	"github.com/Nextdoor/conductor/shared/types"
)

func AllJobsComplete(phaseType types.PhaseType, completedJobs []string) bool {
	expectedJobs := types.JobsForPhase(phaseType)

	if completedJobs == nil && expectedJobs == nil {
		return true
	}

	if completedJobs == nil || expectedJobs == nil {
		return false
	}

	if len(completedJobs) != len(expectedJobs) {
		return false
	}

	sort.Strings(completedJobs)
	sort.Strings(expectedJobs)

	for i := range completedJobs {
		if completedJobs[i] != expectedJobs[i] {
			return false
		}
	}

	return true
}
