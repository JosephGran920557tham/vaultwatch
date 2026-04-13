// Package notify provides integrations for delivering VaultWatch alerts
// to external channels such as Slack and generic HTTP webhooks.
//
// Each notifier in this package implements the alert.Notifier interface,
// allowing it to be registered with the alert.Dispatcher.
//
// Available notifiers:
//
//   - WebhookNotifier: sends alerts as JSON payloads to any HTTP endpoint.
//   - SlackNotifier:   sends formatted messages to a Slack incoming webhook.
//
// Example usage:
//
//	slack, err := notify.NewSlackNotifier(os.Getenv("SLACK_WEBHOOK_URL"), 0)
//	if err != nil {
//		log.Fatal(err)
//	}
//	dispatcher.Register(slack)
package notify
