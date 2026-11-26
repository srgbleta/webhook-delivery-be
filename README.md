# Webhook Delivery Service

A webhook delivery service built with Go and PostgreSQL. Developers register event names and target URLs, then submit events through an API. The service forwards each event’s payload to all matching targets via HTTP POST.

**Work in progress.** Event storage and background delivery are implemented. Retries and authentication are next.

## Current features

- Register webhooks with an event name and target URL.
- View registered webhooks in the dashboard.
- Submit events through the API or use the dashboard’s Test button with a demo payload.
- Save each event and its matching deliveries in a database transaction.
- Respond with `202 Accepted` after saving, without waiting for delivery.
- Process pending deliveries concurrently in a background worker.
- Store each delivery’s status: `pending`, `in_progress`, `done`, or `failed`.

The worker runs alongside the API in the same process and checks for pending deliveries every four seconds. Each delivery is attempted once.

## Next steps

- Retry failed deliveries with a delay and an attempt limit.
- Recover deliveries left `in_progress` after a restart.
- Claim deliveries atomically to prevent workers from picking up the same work.
- Limit concurrent deliveries and handle shutdown gracefully.
- Show delivery results and attempt history in the dashboard.
- Add API keys and ownership checks.
