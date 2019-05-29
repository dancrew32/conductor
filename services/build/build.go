/* Handles building jobs remotely (like Jenkins). */
package build

import "github.com/Nextdoor/conductor/shared/types"

type Service interface {
	TriggerJob(job *types.Job, params map[string]string)
}
