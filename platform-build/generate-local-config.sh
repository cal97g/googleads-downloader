#!/bin/sh

set -eu

echo "\

output_dir: /data

template_vars:
  start_date: $AUDIT_RANGE_START_DATE
  end_date: $AUDIT_RANGE_END_DATE
  end_date_minus_30d: '{{DateSubDays(end_date,30)}}'
  end_date_minus_32d: '{{DateSubDays(end_date,32)}}'
  end_date_minus_90d: '{{DateSubDays(end_date,90)}}'
  end_date_minus_365d: '{{DateSubDays(end_date,365)}}'
  end_date_minus_5y: '{{DateSubDays(end_date,1825)}}'
  perf_initial_date: '{{Earliest(start_date, end_date_minus_32d)}}'

access:
  client_id: $CLIENT_ID
  client_secret: $CLIENT_SECRET
  developer_token: $DEVELOPER_TOKEN
  refresh_token: $REFRESH_TOKEN

$ACCOUNTS
"
