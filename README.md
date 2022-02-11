# lancache-stats
Lancache logfile summarizer, serving stats over HTTP

## Installation
`lancache-stats` is designed to be ran alongside your existing lancache installation. The easiest method to enable this
is to simply add `lancache-stats` to your `docker-compose.yml` from an existing lancache installation, like so:

```
stats:
   image: denverquane/lancache-stats
   env_file: .env
   restart: always
   ports:
     - 5000:5000 # replace left port with whatever port you want
   volumes:
     - ${CACHE_ROOT}/logs:/data/logs:ro
```

By default, this will begin serving stats on port 5000. See [Usage](Usage) for more details on how to