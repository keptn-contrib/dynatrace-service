package adapter

// EventContentAdapter allows to retrieve common data fields of an event
type EventContentAdapter interface {
	GetShKeptnContext() string
	GetEvent() string
	GetSource() string

	GetProject() string
	GetStage() string
	GetService() string
	GetDeployment() string
	GetTestStrategy() string
	GetDeploymentStrategy() string

	GetImage() string
	GetTag() string

	GetLabels() map[string]string
}
