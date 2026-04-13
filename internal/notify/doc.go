// Package notify provides additional notification backends for vaultwatch alerts.
//
// Backends in this package complement the console notifier in the alert package
// by delivering alerts to external systems.
//
// Available notifiers:
//
//   - WebhookNotifier: POSTs a JSON payload to a configurable HTTP endpoint.
//     Suitable for integrating with Slack incoming webhooks, PagerDuty, or any
//     custom HTTP receiver.
//
// All notifiers satisfy the alert.Notifier interface and can be registered with
// an alert.Dispatcher to participate in the standard dispatch pipeline.
//
// Example:
//
//	wh := notify.NewWebhookNotifier("https://hooks.example.com/vault", 10*time.Second)
//	dispatcher.Register(wh)
package notify
