package keptn

import (
	"fmt"
	"github.com/keptn-contrib/dynatrace-service/internal/common"
	"github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
)

type ProjectClientInterface interface {
	AssertProjectExists(projectName string) error
}

type ProjectClient struct {
	client *keptnapi.ProjectHandler
}

func NewDefaultProjectClient() *ProjectClient {
	return NewProjectClient(
		keptnapi.NewProjectHandler(common.GetShipyardControllerURL()))
}

func NewProjectClient(client *keptnapi.ProjectHandler) *ProjectClient {
	return &ProjectClient{
		client: client,
	}
}

// AssertProjectExists returns an error if the project could not be found or retrieved, nil in case of success
func (c *ProjectClient) AssertProjectExists(projectName string) error {
	project, err := c.client.GetProject(
		models.Project{
			ProjectName: projectName,
		})

	if err != nil {
		if err.Code == 404 {
			return fmt.Errorf("project %s does not exist", projectName)
		}

		return fmt.Errorf("could not get project %s: %s", projectName, err.GetMessage())
	}

	if project == nil {
		return fmt.Errorf("project %s is empty", projectName)
	}

	return nil
}
