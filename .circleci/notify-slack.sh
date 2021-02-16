#!/bin/bash

# Posts Slack formatted message to webhook
# requires $SLACK_WEBHOOK, $name, $version and $env.

set -e

for var in SLACK_WEBHOOK name version env url; do
    if [ -z "${!var}" ] ; then
        echo "$var is not set, skipping Slack notification"
        exit 1
    fi
done

fallback="Deployed: $name version $version ($env)"

fields="{\"title\": \"Version\", \"value\": \"$version\", \"short\": true}, {\"title\": \"Environment\", \"value\": \"$env\", \"short\": true}, {\"title\": \"URL\", \"value\": \"$url\", \"short\": true}"

color="#cccccc"
if [[ $env = "Staging" ]]; then
  # amber
  color="#FFBD2E"
elif [[ $env = "Prod" ]]; then
  # green
  color="#16AF59"
fi

json="{\"attachments\": [{\"title\": \"Name: $name\", \"fields\": [$fields], \"fallback\": \"$fallback\", \"color\": \"$color\"}]}"

curl -d "payload=$json" "$SLACK_WEBHOOK"

echo "Slack notification sent"
exit 0
