# adwords-downloader

## Testing and Linting

```shell
$ make
```

## Building

```shell
$ make build
```

## Running

```shell
$ cat local-data/config.yml
output_dir: ./output

access:
  client_id: XX
  client_secret: XX
  developer_token: XX
  refresh_token: XX

template_vars:
  # ISO 8601(YYYY-MM-DD) format.
  start_date: 2019-08-01
  end_date: 2019-08-31
  some_var: value-1
  end_date-90d: '{{DateSubDays(end_date,90)}}'

accounts:
  mcc:
    -
      mcc_id: 7995004967
      account_ids:
        - 4819161749


$ mkdir output
$ adwords-downloader -config local-data/config.yml -profile profiles/search.yml
```

## Platform Integration

Platform integration is taken care of by circleCI making use of the scripts in `platform-build/`
