# Webhook Delivery Service

A webhook delivery service built with Go and PostgreSQL. Developers register event names and target URLs, then submit events through an API. The service forwards each event’s payload to all matching targets via HTTP POST.

**Work in progress.** I’m starting with registering webhooks and sending events. Later, I’ll add queues, retries for failed requests, and delivery history.
