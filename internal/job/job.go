package job

// The different states that a job can be in.
const (
	StateNew        = "new"
	StateQueued     = "queued"
	StateInProgress = "in progress"
	StateComplete   = "complete"
	StateError      = "error"
	StateUnknown    = "?"
)

// The following states are only used by RDC.
const (
	StatePassed = "passed"
	StateFailed = "failed"
)

// DoneStates represents states that a job doesn't transition out of, i.e. once the job is in one of these states,
// it's done.
var DoneStates = []string{StateComplete, StateError, StatePassed, StateFailed}

// Job represents test details and metadata of a test run (aka Job), that is usually associated with a particular test
// execution instance (e.g. VM).
type Job struct {
	ID                  string `json:"id"`
	Passed              bool   `json:"passed"`
	Status              string `json:"status"`
	Error               string `json:"error"`
	BrowserShortVersion string `json:"browser_short_version"`
	BaseConfig          struct {
		PlatformName    string `json:"platformName"`
		PlatformVersion string `json:"platformVersion"`
		DeviceName      string `json:"deviceName"`
	} `json:"base_config"`

	// IsRDC flags a job started as an RDC run.
	IsRDC bool `json:"-"`

	// TimedOut flags a job as an unfinished one.
	TimedOut bool `json:"-"`
}

// TotalStatus returns the total status of a job, combining the result of fields Status + Passed.
func (j Job) TotalStatus() string {
	if Done(j.Status) {
		if j.Passed {
			return StatePassed
		}
		return StateFailed
	}

	return j.Status
}

// Done returns true if the job status is one of DoneStates. False otherwise.
func Done(status string) bool {
	for _, s := range DoneStates {
		if s == status {
			return true
		}
	}

	return false
}

// Service represents the interface for Job interactions.
type Service interface {
	Starter
	Reader
	Writer
	Stopper
	ArtifactDownloader
}

// ArtifactDownloader represents the interface for downloading artifacts.
type ArtifactDownloader interface {
	DownloadArtifact(jobID, suiteName string, realDevice bool) []string
}
