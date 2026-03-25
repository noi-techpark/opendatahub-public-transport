<!--
SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>

SPDX-License-Identifier: CC0-1.0
-->

# SIRI Feed Poller

Polls multiple SIRI endpoints on individual cron schedules and publishes
the raw responses to RabbitMQ. Feed configuration is driven by a YAML
config file mounted via Kubernetes ConfigMap.

For configuration details, see the `.env.example` file and
`infrastructure/feed_config/siri-feeds.yaml`.
