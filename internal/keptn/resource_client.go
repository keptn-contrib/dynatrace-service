package keptn

import (
	"context"
	"errors"
	"fmt"
	"strings"

	keptnmodels "github.com/keptn/go-utils/pkg/api/models"
	v2 "github.com/keptn/go-utils/pkg/api/utils/v2"
	log "github.com/sirupsen/logrus"
)

const projectFieldName = "project"
const stageFieldName = "stage"
const serviceFieldName = "service"
const resourceURIFieldName = "resourceURI"

// ResourceClientInterface defines the methods for interacting with resources of Keptn's configuration service.
type ResourceClientInterface interface {
	// GetResource tries to find the first instance of a given resource on service, stage or project level.
	GetResource(ctx context.Context, project string, stage string, service string, resourceURI string) (string, error)

	// GetProjectResource tries to retrieve a resource on project level.
	GetProjectResource(ctx context.Context, project string, resourceURI string) (string, error)

	// GetStageResource tries to retrieve a resource on stage level.
	GetStageResource(ctx context.Context, project string, stage string, resourceURI string) (string, error)

	// GetServiceResource tries to retrieve a resource on service level.
	GetServiceResource(ctx context.Context, project string, stage string, service string, resourceURI string) (string, error)

	// UploadResource tries to upload a resource.
	UploadResource(ctx context.Context, contentToUpload []byte, remoteResourceURI string, project string, stage string, service string) error
}

// ResourceError represents an error for a resource that was not found.
type ResourceError struct {
	uri     string
	project string
	stage   string
	service string
}

// ResourceNotFoundError represents an error for a resource that was not found.
type ResourceNotFoundError ResourceError

// Error returns a string representation of this error
func (e *ResourceNotFoundError) Error() string {
	return fmt.Sprintf("could not find resource: '%s' %s", e.uri, getLocation(e.service, e.stage, e.project))
}

// ResourceEmptyError represents an error for a resource that was found, but is empty.
type ResourceEmptyError ResourceError

// Error returns a string representation of this error
func (e *ResourceEmptyError) Error() string {
	return fmt.Sprintf("found resource: '%s' %s, but it is empty", e.uri, getLocation(e.service, e.stage, e.project))
}

// ResourceUploadFailedError represents an error for a resource that could not be uploaded.
type ResourceUploadFailedError struct {
	ResourceError
	message string
}

// Error returns a string representation of this error.
func (e *ResourceUploadFailedError) Error() string {
	return fmt.Sprintf("could not upload resource: '%s' %s: %s", e.uri, getLocation(e.service, e.stage, e.project), e.message)
}

// ResourceRetrievalFailedError represents an error for a resource that could not be retrieved because of an error.
type ResourceRetrievalFailedError struct {
	ResourceError
	cause error
}

// Error returns a string representation of this error.
func (e *ResourceRetrievalFailedError) Error() string {
	return fmt.Sprintf("could not retrieve resource: '%s' %s: %v", e.uri, getLocation(e.service, e.stage, e.project), e.cause)
}

// Unwrap returns the cause of the ResourceRetrievalFailedError.
func (e *ResourceRetrievalFailedError) Unwrap() error {
	return e.cause
}

func getLocation(service string, stage string, project string) string {
	var location string

	if service != "" {
		location += fmt.Sprintf(" for service '%s'", service)
	}
	if stage != "" {
		location += fmt.Sprintf(" at stage '%s'", stage)
	}
	if project != "" {
		location += fmt.Sprintf(" of project '%s'", project)
	}

	return strings.TrimLeft(location, " ")
}

// ResourceClient is the default implementation for the ResourceClientInterface using a Keptn api.ResourcesV1Interface.
type ResourceClient struct {
	client v2.ResourcesInterface
}

// NewResourceClient creates a new ResourceClient using a api.ResourcesV1Interface.
func NewResourceClient(client v2.ResourcesInterface) *ResourceClient {
	return &ResourceClient{
		client: client,
	}
}

// GetResource tries to find the first instance of a given resource on service, stage or project level.
func (rc *ResourceClient) GetResource(ctx context.Context, project string, stage string, service string, resourceURI string) (string, error) {
	var rnfErrorType *ResourceNotFoundError
	if project != "" && stage != "" && service != "" {
		keptnResourceContent, err := rc.GetServiceResource(ctx, project, stage, service, resourceURI)
		if errors.As(err, &rnfErrorType) {
			log.WithFields(
				log.Fields{
					projectFieldName:     project,
					stageFieldName:       stage,
					serviceFieldName:     service,
					resourceURIFieldName: resourceURI,
				}).Debug("Resource not available for service")
		} else if err != nil {
			return "", err
		} else {
			log.WithFields(
				log.Fields{
					projectFieldName:     project,
					stageFieldName:       stage,
					serviceFieldName:     service,
					resourceURIFieldName: resourceURI,
				}).Info("Found resource for service")
			return keptnResourceContent, nil
		}
	}

	if project != "" && stage != "" {
		keptnResourceContent, err := rc.GetStageResource(ctx, project, stage, resourceURI)
		if errors.As(err, &rnfErrorType) {
			log.WithFields(
				log.Fields{
					projectFieldName:     project,
					stageFieldName:       stage,
					resourceURIFieldName: resourceURI,
				}).Debug("Resource not available for stage")
		} else if err != nil {
			return "", err
		} else {
			log.WithFields(
				log.Fields{
					projectFieldName:     project,
					stageFieldName:       stage,
					resourceURIFieldName: resourceURI,
				}).Info("Found resource for stage")
			return keptnResourceContent, nil
		}
	}

	if project != "" {
		keptnResourceContent, err := rc.GetProjectResource(ctx, project, resourceURI)
		if errors.As(err, &rnfErrorType) {
			log.WithFields(
				log.Fields{
					projectFieldName:     project,
					resourceURIFieldName: resourceURI,
				}).Debug("Resource not available for project")
		} else if err != nil {
			return "", err
		} else {
			log.WithFields(
				log.Fields{projectFieldName: project,
					resourceURIFieldName: resourceURI,
				}).Info("Found resource for project")
			return keptnResourceContent, nil
		}
	}

	log.WithField(resourceURIFieldName, resourceURI).Info("Resource not found")
	return "", &ResourceNotFoundError{uri: resourceURI, project: project, stage: stage, service: service}
}

// GetServiceResource tries to retrieve a resource on service level.
func (rc *ResourceClient) GetServiceResource(ctx context.Context, project string, stage string, service string, resourceURI string) (string, error) {
	return getResourceByFunc(
		func() (*keptnmodels.Resource, error) {
			return rc.client.GetResource(ctx,
				*v2.NewResourceScope().Project(project).Stage(stage).Service(service).Resource(resourceURI),
				v2.ResourcesGetResourceOptions{})
		},
		func() *ResourceNotFoundError {
			return &ResourceNotFoundError{uri: resourceURI, project: project, stage: stage, service: service}
		},
		func(cause error) *ResourceRetrievalFailedError {
			return &ResourceRetrievalFailedError{ResourceError{uri: resourceURI, project: project, stage: stage, service: service}, cause}
		},
		func() *ResourceEmptyError {
			return &ResourceEmptyError{uri: resourceURI, project: project, stage: stage, service: service}
		})
}

// GetStageResource tries to retrieve a resource on stage level.
func (rc *ResourceClient) GetStageResource(ctx context.Context, project string, stage string, resourceURI string) (string, error) {
	return getResourceByFunc(
		func() (*keptnmodels.Resource, error) {
			return rc.client.GetResource(ctx,
				*v2.NewResourceScope().Project(project).Stage(stage).Resource(resourceURI),
				v2.ResourcesGetResourceOptions{})
		},
		func() *ResourceNotFoundError {
			return &ResourceNotFoundError{uri: resourceURI, project: project, stage: stage}
		},
		func(cause error) *ResourceRetrievalFailedError {
			return &ResourceRetrievalFailedError{ResourceError{uri: resourceURI, project: project, stage: stage}, cause}
		},
		func() *ResourceEmptyError {
			return &ResourceEmptyError{uri: resourceURI, project: project, stage: stage}
		})
}

// GetProjectResource tries to retrieve a resource on project level.
func (rc *ResourceClient) GetProjectResource(ctx context.Context, project string, resourceURI string) (string, error) {
	return getResourceByFunc(
		func() (*keptnmodels.Resource, error) {
			return rc.client.GetResource(ctx,
				*v2.NewResourceScope().Project(project).Resource(resourceURI),
				v2.ResourcesGetResourceOptions{})
		},
		func() *ResourceNotFoundError { return &ResourceNotFoundError{uri: resourceURI, project: project} },
		func(cause error) *ResourceRetrievalFailedError {
			return &ResourceRetrievalFailedError{ResourceError{uri: resourceURI, project: project}, cause}
		},
		func() *ResourceEmptyError { return &ResourceEmptyError{uri: resourceURI, project: project} })
}

func getResourceByFunc(
	resFunc func() (*keptnmodels.Resource, error),
	rnfErrFunc func() *ResourceNotFoundError,
	rrfErrFunc func(cause error) *ResourceRetrievalFailedError,
	reErrFunc func() *ResourceEmptyError) (string, error) {
	resource, err := resFunc()
	if err != nil {
		if err == v2.ResourceNotFoundError {
			return "", rnfErrFunc()
		}

		return "", rrfErrFunc(err)
	}
	if resource.ResourceContent == "" {
		return "", reErrFunc()
	}

	return resource.ResourceContent, nil
}

// UploadResource tries to upload a resource.
func (rc *ResourceClient) UploadResource(ctx context.Context, contentToUpload []byte, remoteResourceURI string, project string, stage string, service string) error {
	resources := []*keptnmodels.Resource{{ResourceContent: string(contentToUpload), ResourceURI: &remoteResourceURI}}
	_, err := rc.client.CreateResources(ctx, project, stage, service, resources, v2.ResourcesCreateResourcesOptions{})
	if err != nil {
		return &ResourceUploadFailedError{
			ResourceError{
				uri:     remoteResourceURI,
				project: project,
				stage:   stage,
				service: service,
			},
			err.GetMessage(),
		}
	}

	log.WithField("remoteResourceURI", remoteResourceURI).Info("Uploaded file")
	return nil
}
