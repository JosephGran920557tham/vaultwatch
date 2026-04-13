// Package notify provides implementations of the alert.Notifier interface
// for delivering VaultWatch lease expiration alerts through various channels.
//
// Supported notifiers:
//
//   - WebhookNotifier: sends alerts as JSON payloads to an HTTP endpoint
//   - SlackNotifier:   sends formatted messages to a Slack incoming webhook
//   - EmailNotifier:   delivers alerts via SMTP email
//   - PagerDutyNotifier: triggers PagerDuty incidents via Events API v2
//
// Each notifier is constructed with channel-specific configuration and
// implements the Send(alert.Alert) error method. Notifiers are registered
// with an alert.Dispatcher which handles severity filtering and fan-out.
//
// Example usage:
//
//	pd, err := notify.NewPagerDutyNotifier(os.Getenv("PD_KEY"), 10*time.Second)
//	if err != nil {
//		log.Fatal(err)
//	}
//	dispatcher.Register(pd)
package notify
