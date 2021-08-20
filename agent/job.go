package agent

type ErrJobAlreadyStarted error

// Job represents a piece of work for the agent to perform.
type Job interface {
	// Claim confers exclusive rights on the agent to process the job,
	// preventing any other agent from processing the job. ErrJobAlreadyStarted
	// should be returned if another agent is performing the job.
	Claim() error
	// Process actually carries out the piece-of-work the job represents
	Process() error
}

type Plan struct{}

func (p *Plan) Claim() error {
	// call run#start-plan

	return nil
}
