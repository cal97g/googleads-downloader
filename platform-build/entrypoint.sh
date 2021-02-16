#!/bin/bash

set -euo  pipefail

removequotes() {
    sed -e 's/^"//' -e 's/"$//' <<< "$1"
}


printf -- "-----------------------------------\n"
printf "Adwords Downloader\n"
printf "TASK ID: $TASK_ID\n"
printf "Audit Date: $AUDIT_RANGE_START_DATE - $AUDIT_RANGE_END_DATE\n"
printf -- "-----------------------------------\n"

# read client id/secret and developer token from SSM parameter store
CLIENT_ID=$(aws ssm get-parameters --with-decryption --names /$PERCEPT_STAGE/adwords_client_id --query "Parameters[0].Value")
export CLIENT_ID=$(removequotes $CLIENT_ID)

CLIENT_SECRET=$(aws ssm get-parameters --with-decryption --names /$PERCEPT_STAGE/adwords_client_secret --query "Parameters[0].Value")
export CLIENT_SECRET=$(removequotes $CLIENT_SECRET)

DEVELOPER_TOKEN=$(aws ssm get-parameters --with-decryption --names adwords_developer_token_1 --query "Parameters[0].Value")
export DEVELOPER_TOKEN=$(removequotes $DEVELOPER_TOKEN)

# dir to hold series config and generated downloader config
mkdir config

# download series config from S3 (a snapshot created for this task)
aws s3 cp "s3://$S3_CONFIG_SOURCE/$TASK_ID/config.json" config/platform_config.json --quiet

printf "Series config: $(cat config/platform_config.json)\n"

# transform series config into yml format
yq r config/platform_config.json > config/platform_config.yml

# extract downloader key from series config
export ACCOUNTS=$(yq r config/platform_config.yml downloader)

# use env vars to generate downloader config
./generate-local-config.sh > config/local_config.yml

# output dir
mkdir /data

# map $PROFILE to downloader profile files
seriesprofile="${PROFILE:-enabled_only}"
case "$seriesprofile" in
"in_scope" | "inscope_only" | "vf_50")
    profilepath=./profiles/current/combined/search.inscope.yml
    ;;
"default" | "enabled_only" | "effectively_enabled")
    profilepath=./profiles/current/combined/search.ee.yml
    ;;
*)
    printf "unknown profile, using default (enabled_only)"
    profilepath=./profiles/current/combined/search.ee.yml
    ;;
esac

apiversion="v5"

# execute downloader bin
./adwords-downloader -api-version $apiversion -config config/local_config.yml -profile $profilepath --verbose --compress

# compress
tar -czf output.tar.gz /data

# upload output dir to S3
aws s3 cp output.tar.gz "s3://$S3_EXPORT_TARGET/$TASK_ID/output.tar.gz"

