package adapter

type EventAdapter interface {
	GetContext() string
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
