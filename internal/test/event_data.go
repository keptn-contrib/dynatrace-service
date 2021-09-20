package test

type EventData struct {
	Context string
	Source  string
	Event   string

	Project            string
	Stage              string
	Service            string
	Deployment         string
	TestStrategy       string
	DeploymentStrategy string

	Labels map[string]string
}

// GetShKeptnContext returns the shkeptncontext
func (e *EventData) GetShKeptnContext() string {
	return e.Context
}

// GetSource returns the source specified in the CloudEvent context
func (e *EventData) GetSource() string {
	return e.Source
}

// GetEvent returns the event type
func (e *EventData) GetEvent() string {
	return e.Event
}

// GetProject returns the project
func (e *EventData) GetProject() string {
	return e.Project
}

// GetStage returns the stage
func (e *EventData) GetStage() string {
	return e.Stage
}

// GetService returns the service
func (e *EventData) GetService() string {
	return e.Service
}

// GetDeployment returns the name of the deployment
func (e *EventData) GetDeployment() string {
	return e.Deployment
}

// GetTestStrategy returns the used test strategy
func (e *EventData) GetTestStrategy() string {
	return e.TestStrategy
}

// GetDeploymentStrategy returns the used deployment strategy
func (e EventData) GetDeploymentStrategy() string {
	return e.DeploymentStrategy
}

// GetLabels returns a map of labels
func (e *EventData) GetLabels() map[string]string {
	return e.Labels
}
