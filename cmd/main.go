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

	api "github.com/keptn/go-utils/pkg/api/utils"
	eventsource "github.com/keptn/go-utils/pkg/sdk/connector/eventsource/nats"
	nats "github.com/keptn/go-utils/pkg/sdk/connector/nats"

	"github.com/keptn/go-utils/pkg/sdk/connector/controlplane"
	"github.com/keptn/go-utils/pkg/sdk/connector/logforwarder"
	"github.com/keptn/go-utils/pkg/sdk/connector/subscriptionsource"
	"github.com/keptn/go-utils/pkg/sdk/connector/types"

	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn/go-utils/pkg/api/models"
	keptnapi "github.com/keptn/go-utils/pkg/api/utils"
	"github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type dynatraceService struct {
	onEvent func(eventSenderClient *keptn.EventSenderClient, event cloudevents.Event)
}

func main() {
	log.SetLevel(env.GetLogLevel())
	os.Exit(_main())
}

func _main() int {
	// start health endpoint
	// TODO: 2022-06-14: Check: is it possible to terminate liveness cleanly?
	go func() {
		keptnapi.RunHealthEndpoint("8070")
	}()

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

	natsConnector := nats.NewFromEnv()
	controlPlane, err := connectToControlPlane(natsConnector)
	if err != nil {
		log.WithError(err).Fatal("Could not connect to control plane")
	}

	// start readiness endpoint
	// TODO: 2022-06-14: Check: is it possible to terminate readiness cleanly?
	go func() {
		keptnapi.RunHealthEndpoint("8080", keptnapi.WithPath("/ready"), keptnapi.WithReadinessConditionFunc(func() bool {
			return controlPlane.IsRegistered()
		}))
	}()

	// register for events
	// the actual processing is done in a separate goroutine so that it doesn't block other events
	log.Info("Registering with control plane")
	err = controlPlane.Register(notifyCtx, dynatraceService{onEvent: func(eventSenderClient *keptn.EventSenderClient, event cloudevents.Event) {
		workerWaitGroup.Add(1)
		go func() {
			defer workerWaitGroup.Done()
			gotEvent(workCtx, replyCtx, eventSenderClient, event)
		}()
	}})
	if err != nil {
		log.WithError(err).Error("Could not register control plane")
	}

	// wait for all existing events (i.e. worker go routines to finish)
	log.Info("Waiting for existing processing to finish")
	stopNotify()
	workerWaitGroup.Wait()

	// TODO: 2022-07-12: Once available, this should be updated to use a context when flushing the connection.
	err = natsConnector.Disconnect()
	if err != nil {
		log.WithError(err).Error("Could not disconnect NATS connector")
	}

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
	d.onEvent(keptn.NewEventSenderClient(eventSender), cloudEvent)
	return nil
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
			// TODO: fixed to "0.16.0" until Keptn provides a default
			DistributorVersion: "0.16.0",
		},
		Subscriptions: []models.EventSubscription{
			createEventSubscription("sh.keptn.event.monitoring.configure"),
			createEventSubscription("sh.keptn.events.problem"),
			createEventSubscription("sh.keptn.event.action.triggered"),
			createEventSubscription("sh.keptn.event.action.started"),
			createEventSubscription("sh.keptn.event.action.finished"),
			createEventSubscription("sh.keptn.event.get-sli.triggered"),
			createEventSubscription("sh.keptn.event.deployment.finished"),
			createEventSubscription("sh.keptn.event.test.triggered"),
			createEventSubscription("sh.keptn.event.test.finished"),
			createEventSubscription("sh.keptn.event.evaluation.finished"),
			createEventSubscription("sh.keptn.event.release.triggered"),
		},
	}
}

func gotEvent(workCtx context.Context, replyCtx context.Context, eventSender *keptn.EventSenderClient, event cloudevents.Event) {
	clientFactory, err := keptn.NewClientFactory()
	if err != nil {
		log.WithError(err).Error("Could not create a Keptn client factory")
		return
	}

	handler, err := event_handler.NewEventHandler(workCtx, clientFactory, eventSender, event)
	if err != nil {
		log.WithError(err).Error("NewEventHandler() returned an error")
		return
	}

	err = handler.HandleEvent(workCtx, replyCtx)
	if err != nil {
		log.WithError(err).Error("HandleEvent() returned an error")
	}
}

func connectToControlPlane(natsConnector *nats.NatsConnector) (*controlplane.ControlPlane, error) {
	apiSet, err := api.NewInternal(&http.Client{}, keptn.GetV1InClusterAPIMappings())
	if err != nil {
		return nil, fmt.Errorf("could not create internal Keptn API set: %w", err)
	}

	return controlplane.New(
		subscriptionsource.New(apiSet.UniformV1()),
		eventsource.New(natsConnector),
		logforwarder.New(apiSet.LogsV1())), nil
}

func createEventSubscription(event string) models.EventSubscription {
	return models.EventSubscription{
		Event:  event,
		Filter: models.EventSubscriptionFilter{},
	}
}
