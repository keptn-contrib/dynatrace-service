package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	context2 "github.com/keptn-contrib/dynatrace-service/internal/context"
	"github.com/keptn-contrib/dynatrace-service/internal/env"
	"github.com/keptn-contrib/dynatrace-service/internal/event_handler"
	"github.com/keptn-contrib/dynatrace-service/internal/keptn"
	"github.com/keptn-contrib/dynatrace-service/internal/onboard"

	"github.com/keptn/keptn/cp-common/api"
	"github.com/keptn/keptn/cp-connector/pkg/controlplane"
	"github.com/keptn/keptn/cp-connector/pkg/eventsource"
	"github.com/keptn/keptn/cp-connector/pkg/logforwarder"
	"github.com/keptn/keptn/cp-connector/pkg/nats"
	"github.com/keptn/keptn/cp-connector/pkg/subscriptionsource"
	"github.com/keptn/keptn/cp-connector/pkg/types"

	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	"github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type dynatraceService struct {
	onEvent func(keptnClient *keptn.Client, event cloudevents.Event) error
}

func main() {
	log.SetLevel(env.GetLogLevel())
	os.Exit(_main())
}

func _main() int {
	// root context
	ctx := context.Background()

	// notifyCtx is done when a termination signal is received
	notifyCtx, stopNotify := signal.NotifyContext(ctx,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer stopNotify()

	// workCtx will be cancelled after the grace period after notifyCtx is done
	workCtx, stopWorkPeriod := context2.NewTriggeredTimeoutContext(ctx, notifyCtx, env.GetWorkGracePeriod(), "workCtx")
	defer stopWorkPeriod()

	// replyCtx will be cancelled after a cleanup period after workCtx is done
	replyCtx, stopReplyPeriod := context2.NewTriggeredTimeoutContext(ctx, workCtx, env.GetReplyGracePeriod(), "replyCtx")
	defer stopReplyPeriod()

	workerWaitGroup := &sync.WaitGroup{}
	if env.IsServiceSyncEnabled() {
		workerWaitGroup.Add(1)
		go func() {
			defer workerWaitGroup.Done()
			serviceSynchronizer, err := onboard.NewDefaultServiceSynchronizer()
			if err != nil {
				log.WithError(err).Error("Could not create service synchronizer")
				return
			}

			serviceSynchronizer.Run(notifyCtx, workCtx)
		}()
	}

	controlPlane, err := connectToControlPlane()
	if err != nil {
		log.WithError(err).Fatal("Could not connect to control plane")
	}

	go func() {
		// TODO: fix port?
		keptnapi.RunHealthEndpoint("8080", keptnapi.WithReadinessConditionFunc(func() bool {
			return controlPlane.IsRegistered()
		}))
	}()

	// register for events
	// the actual processing is done in a separate goroutine so that it doesn't block other events
	log.Info("Registering with control plane")
	controlPlane.Register(notifyCtx, dynatraceService{onEvent: func(keptnClient *keptn.Client, event cloudevents.Event) error {
		workerWaitGroup.Add(1)
		go func() {
			defer workerWaitGroup.Done()
			gotEvent(workCtx, replyCtx, keptnClient, event)
		}()
		return nil
	}})

	// wait for all existing events (i.e. worker go routines to finish)
	log.Info("Waiting for existing processing to finish")
	stopNotify()
	workerWaitGroup.Wait()
	stopWorkPeriod()
	stopReplyPeriod()

	log.Info("Shutdown complete")
	return 0
}

// OnEvent is called when a new event was received.
func (d dynatraceService) OnEvent(ctx context.Context, event models.KeptnContextExtendedCE) error {
	eventSender, ok := ctx.Value(types.EventSenderKey).(controlplane.EventSender)
	if !ok {
		return fmt.Errorf("could not get eventSender from context")
	}

	cloudEvent := v0_2_0.ToCloudEvent(event)
	keptnClient, err := keptn.NewClient(&cpEventSender{sender: eventSender}, cloudEvent)
	if err != nil {
		return err
	}

	return d.onEvent(keptnClient, cloudEvent)
}

// RegistrationData is called to get the initial registration data.
func (d dynatraceService) RegistrationData() controlplane.RegistrationData {
	metadata, err := env.GetK8sMetadata()
	if err != nil {
		log.WithError(err).Fatal()
	}

	return controlplane.RegistrationData{
		Name: metadata.DeploymentName(),
		MetaData: models.MetaData{
			Hostname:           metadata.NodeName(),
			IntegrationVersion: metadata.DeploymentVersion(),
			Location:           metadata.DeploymentComponent(),
			KubernetesMetaData: models.KubernetesMetaData{
				Namespace:      metadata.Namespace(),
				PodName:        metadata.PodName(),
				DeploymentName: metadata.DeploymentName(),
			},
		},
		Subscriptions: []models.EventSubscription{
			{
				Event:  "sh.keptn.event.>",
				Filter: models.EventSubscriptionFilter{},
			},
		},
	}
}

func gotEvent(workCtx context.Context, replyCtx context.Context, keptnClient *keptn.Client, event cloudevents.Event) {
	handler, err := event_handler.NewEventHandler(workCtx, keptnClient, event)
	if err != nil {
		log.WithError(err).Error("NewEventHandler() returned an error")
		return
	}

	handler.HandleEvent(workCtx, replyCtx)
	if err != nil {
		log.WithError(err).Error("HandleEvent() returned an error")
	}
}

func connectToControlPlane() (*controlplane.ControlPlane, error) {
	apiSet, err := api.NewInternal(&http.Client{}, keptn.GetInClusterAPIMappings())
	if err != nil {
		return nil, fmt.Errorf("could not create internal Keptn API set: %w", err)
	}

	natsConnector, err := nats.ConnectFromEnv()
	if err != nil {
		return nil, fmt.Errorf("could not connect to NATS: %w", err)
	}

	return controlplane.New(
		subscriptionsource.New(apiSet.UniformV1()),
		eventsource.New(natsConnector),
		logforwarder.New(apiSet.LogsV1())), nil
}

type cpEventSender struct {
	sender controlplane.EventSender
}

// SendEvent sends a cloud event.
func (e *cpEventSender) SendEvent(event cloudevents.Event) error {
	return e.Send(context.TODO(), event)
}

// Send sends a cloud event.
func (e *cpEventSender) Send(ctx context.Context, event cloudevents.Event) error {
	keptnEvent, err := v0_2_0.ToKeptnEvent(event)
	if err != nil {
		return err
	}
	return e.sender(keptnEvent)
}
