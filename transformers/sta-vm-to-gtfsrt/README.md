<!--
SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>

SPDX-License-Identifier: CC0-1.0
-->

# STA Vehicle Monitoring to GTFS-RT

Listens on a RabbitMQ queue for SIRI Lite Vehicle Monitoring data from STA
and converts it to GTFS-RT VehiclePositions using NeTEx and GTFS static data
for consistent ID mapping.

Downloads NeTEx and GTFS static data at startup and refreshes every 24 hours.
Outputs both protobuf (.pb) and JSON (.json) to the fileserver.
