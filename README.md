# Webhook Delivery Service

A webhook delivery service built with Go and PostgreSQL. Developers register event names and target URLs, then submit events through an API. The service forwards each event’s payload to all matching targets via HTTP POST.

## V0 — Complete

The initial end-to-end flow is implemented and verified:

- Register a webhook with an event name and target URL.
- View registered webhooks in the dashboard.
- Submit an event with an event name and payload through the API (or click the Test button next to each webhook entry, it will use a demo payload in such case).
- The service finds all webhooks matching that event name and sends the payload to their target URLs concurrently, waiting for all delivery attempts to finish before responding.

V0 attempts each delivery once. It does not yet queue deliveries, retry failed requests, or store delivery history.

## Next steps

- Store delivery results
- Move delivery into a background worker
- Add retries
- Add API keys and ownership
