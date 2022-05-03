package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/keptn-contrib/dynatrace-service/internal/env"
	"github.com/keptn-contrib/dynatrace-service/internal/event_handler"
	"github.com/keptn-contrib/dynatrace-service/internal/health"
	"github.com/keptn-contrib/dynatrace-service/internal/onboard"

	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port       int    `envconfig:"RCV_PORT" default:"8080"`
	Path       string `envconfig:"RCV_PATH" default:"/"`
	HealthPort int    `envconfig:"HEALTH_PORT" default:"8070"`
}

func main() {
	log.SetLevel(env.GetLogLevel())

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.WithError(err).Fatal("Failed to process env var")
	}

	os.Exit(_main(env))
}

func _main(envCfg envConfig) int {

	healthEndpoint := health.NewHealthEndpoint(fmt.Sprintf(":%d", envCfg.HealthPort))
	healthEndpoint.Start()

	// root context
	ctx := cloudevents.WithEncodingStructured(context.Background())

	// notifyCtx is done when a termination signal is received
	notifyCtx, stopNotify := signal.NotifyContext(ctx,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer stopNotify()

	// workCtx will be cancelled after the grace period after notifyCtx is done
	workCtx, cancelWorkCtx := context.WithCancel(ctx)

	// replyCtx will be cancelled after a cleanup period after workCtx is done
	replyCtx, cancelReplyCtx := context.WithCancel(ctx)

	// start helper go routine to provide cancellation after grace period after signal
	helperWaitGroup := &sync.WaitGroup{}
	helperWaitGroup.Add(1)
	go func() {
		defer helperWaitGroup.Done()
		defer cancelWorkCtx()
		defer cancelReplyCtx()

		// wait for notify signal
		<-notifyCtx.Done()

		// calculate effective grace period, allow 5 seconds of slack time
		const slackSeconds = 5
		const desiredCleanupPeriodSeconds = 5

		actualCleanupPeriodSeconds := env.GetGracePeriodSeconds() - slackSeconds
		if actualCleanupPeriodSeconds > desiredCleanupPeriodSeconds {
			actualCleanupPeriodSeconds = desiredCleanupPeriodSeconds
		} else if actualCleanupPeriodSeconds < 0 {
			actualCleanupPeriodSeconds = 0
		}

		actualGracePeriodSeconds := env.GetGracePeriodSeconds() - actualCleanupPeriodSeconds - slackSeconds
		if actualGracePeriodSeconds < 0 {
			actualGracePeriodSeconds = 0
		}

		log.WithField("actualGracePeriodSeconds", actualGracePeriodSeconds).Info("Notified for shutdown, starting grace period")

		// wait out the grace period but stop if already cancelled, i.e. if workers have already finished
		select {
		case <-replyCtx.Done():
			return
		case <-time.After(time.Duration(actualGracePeriodSeconds) * time.Second):
		}

		log.WithField("actualCleanupPeriodSeconds", actualCleanupPeriodSeconds).Info("Grace period has expired, cancelling all work and starting clean up period")
		cancelWorkCtx()

		// wait out the reply period but stop if already cancelled, i.e. if workers have already finished
		select {
		case <-replyCtx.Done():
			return
		case <-time.After(time.Duration(actualCleanupPeriodSeconds) * time.Second):
		}

		log.Info("Cleanup period has expired, cancelling all cleanup")
	}()

	workerWaitGroup := &sync.WaitGroup{}
	if env.IsServiceSyncEnabled() {
		workerWaitGroup.Add(1)
		go func() {
			defer workerWaitGroup.Done()
			onboard.NewDefaultServiceSynchronizer().Run(notifyCtx, workCtx)
		}()
	}

	log.WithFields(log.Fields{"port": envCfg.Port, "path": envCfg.Path}).Debug("Initializing cloudevents client")
	c, err := cloudevents.NewClientHTTP(cloudevents.WithPath(envCfg.Path), cloudevents.WithPort(envCfg.Port), cloudevents.WithGetHandlerFunc(health.HTTPGetHandler))
	if err != nil {
		log.WithError(err).Fatal("Failed to create client")
	}

	// start actually receiving cloud events
	// the actual processing is done in a separate go routine which receives only the graceful context
	// this allows the incoming cloud event to be acknowledged immediately to avoid hitting a timeout specified in the distributor
	log.Info("Starting receiver")
	err = c.StartReceiver(notifyCtx,
		func(event cloudevents.Event) {
			workerWaitGroup.Add(1)
			go func() {
				defer workerWaitGroup.Done()
				gotEvent(workCtx, replyCtx, event)
			}()
		})

	// at this point receiver has finished, i.e no new cloud events will be received
	if err != nil {
		log.WithError(err).Error("Receiver finished with error")
	}

	// wait for all existing events (i.e. worker go routines to finish)
	log.Info("Waiting for existing processing to finish")
	stopNotify()
	workerWaitGroup.Wait()
	cancelWorkCtx()
	cancelReplyCtx()
	helperWaitGroup.Wait()

	healthEndpoint.Stop()

	log.Info("Shutdown complete")
	return 0
}

func gotEvent(workCtx context.Context, replyCtx context.Context, event cloudevents.Event) {
	err := event_handler.NewEventHandler(workCtx, event).HandleEvent(workCtx)
	if err != nil {
		log.WithError(err).Error("HandleEvent() returned an error")
	}
}
