package main

import (
	"context"
	"github.com/keptn-contrib/dynatrace-service/pkg/lib"
	keptncommon "github.com/keptn/go-utils/pkg/lib/keptn"
	"log"
	"os"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	"github.com/keptn-contrib/dynatrace-service/pkg/event_handler"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int    `envconfig:"RCV_PORT" default:"8080"`
	Path string `envconfig:"RCV_PATH" default:"/"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}

	if common.RunLocal || common.RunLocalTest {
		log.Println("env=runlocal: Running with local filesystem to fetch resources")
	}

	os.Exit(_main(os.Args[1:], env))
}

func _main(args []string, env envConfig) int {

	// check the connection to the Keptn API
	keptnCredentials, err := common.GetKeptnCredentials()
	if err != nil {
		log.Fatalf("failed to get Keptn API credentials: %s", err.Error())
	}

	log.Printf("Verifying access to Keptn API at %s", keptnCredentials.ApiURL)

	err = common.CheckKeptnConnection(keptnCredentials)
	if err != nil {
		log.Printf("Warning: Keptn API connection cannot be verified. This might be due to a no-loopback policy of your LoadBalancer. The endpoint might still be reachable from outside the cluster. %s", err.Error())
	}

	if lib.IsServiceSyncEnabled() {
		lib.ActivateServiceSynchronizer()
	}

	ctx := context.Background()
	ctx = cloudevents.WithEncodingStructured(ctx)

	p, err := cloudevents.NewHTTP(cloudevents.WithPath(env.Path), cloudevents.WithPort(env.Port))
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	log.Fatal(c.StartReceiver(ctx, gotEvent))

	return 0
}

func gotEvent(ctx context.Context, event cloudevents.Event) error {

	var shkeptncontext string
	_ = event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	logger := keptncommon.NewLogger(shkeptncontext, event.Context.GetID(), "dynatrace-service")

	dynatraceEventHandler, err := event_handler.NewEventHandler(event, logger)

	if err != nil {
		return err
	}

	err = dynatraceEventHandler.HandleEvent()

	if err != nil {
		return err
	}

	return nil
}
