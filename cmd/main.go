package main

import (
	"context"
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
	Port int    `envconfig:"RCV_PORT" default:"8080"`
	Path string `envconfig:"RCV_PATH" default:"/"`
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

	// root context
	ctx := cloudevents.WithEncodingStructured(context.Background())

	// notifyCtx is done when a termination signal is received
	notifyCtx, stopNotify := signal.NotifyContext(ctx,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer stopNotify()

	// gracefulCtx will be cancelled after the grace period after notifyContext is done
	gracefulCtx, cancelGracePeriod := context.WithCancel(ctx)

	// start helper go routine to provide cancellation after grace period after signal
	helperWaitGroup := &sync.WaitGroup{}
	helperWaitGroup.Add(1)
	go func() {
		defer helperWaitGroup.Done()
		defer cancelGracePeriod()

		// wait for notify signal
		<-notifyCtx.Done()

		// calculate effective grace period, allow 5 seconds of slack time
		effectiveGracePeriodSeconds := env.GetGracePeriodSeconds() - 5
		if effectiveGracePeriodSeconds <= 0 {
			log.Info("Skipping grace period, cancelling all work")
			return
		}

		log.WithField("effectiveGracePeriodSeconds", effectiveGracePeriodSeconds).Info("Notified for shutdown, starting grace period")

		// wait out the grace period but stop if already cancelled, i.e. if workers have already finished
		select {
		case <-time.After(time.Duration(effectiveGracePeriodSeconds) * time.Second):
			log.Info("Grace period has expired, cancelling all work")
		case <-gracefulCtx.Done():
		}
	}()

	workerWaitGroup := &sync.WaitGroup{}
	if env.IsServiceSyncEnabled() {
		workerWaitGroup.Add(1)
		go func() {
			defer workerWaitGroup.Done()
			onboard.NewDefaultServiceSynchronizer().Run(notifyCtx, gracefulCtx)
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
				gotEvent(gracefulCtx, event)
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
	cancelGracePeriod()
	helperWaitGroup.Wait()

	log.Info("Shutdown complete")
	return 0
}

func gotEvent(ctx context.Context, event cloudevents.Event) {
	err := event_handler.NewEventHandler(ctx, event).HandleEvent(ctx)
	if err != nil {
		log.WithError(err).Error("HandleEvent() returned an error")
	}
}
