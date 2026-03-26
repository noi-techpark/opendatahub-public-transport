<!--
SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>

SPDX-License-Identifier: CC0-1.0
-->

# Feed Fetcher

Generic feed poller for public transport data (SIRI, SIRI Lite, GTFS-RT, etc.).
Polls multiple endpoints on individual cron schedules and publishes the raw
responses to RabbitMQ. Feed configuration is driven by a YAML config file
mounted via Kubernetes ConfigMap.

For configuration details, see the `.env.example` file and
`infrastructure/feed_config/sta-siri-feeds.yaml`.
