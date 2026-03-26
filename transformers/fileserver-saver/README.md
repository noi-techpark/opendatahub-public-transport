<!--
SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>

SPDX-License-Identifier: CC0-1.0
-->

# Fileserver Saver

Listens on multiple RabbitMQ queues and saves raw feed payloads to an nginx
fileserver via HTTP PUT. Sink configuration is driven by a YAML config file
mounted via Kubernetes ConfigMap.

Each config entry maps a queue to a fileserver destination path. The transformer
extracts the raw payload from the feed data and saves it as-is.

For configuration details, see the `.env.example` file and
`infrastructure/sink_config/sta-siri-lite.yaml`.
