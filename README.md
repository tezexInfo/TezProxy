# TezProxy

TezProxy is a simple Go App that servers as a Proxy fo the tezos RPC
that can block ad whitelist specific calls, and utilize a simple LRU-Cache for
improved performance



## Example Config

TezProxy will look for a config file called `config.yml` in either its current directory
or under `/etc/tezproxy/`

Example `config.yml`:

```
tezos:
  host: 127.0.0.1  // Tezos Node Host
  port: 8732       // Tezos Node Port

server:
  port: 8899       // Port that TezProxy will run at

proxy:
  readTimeout: 1
  writeTimeout: 5
  idleTimeout: 120
  whitelistedMethods:   // all allowed calls must be whitelisted
    - /chains/main/blocks(.*?)
  blockedMethods:       // this can blacklist calls that are otherwise 
                        // whitelisted, for finer control of whitelist
    - (.*?)context/contracts$
    - /monitor(.*?)
    - /netowork(.*?)
  dontCache:            // all calls get cached per default, except:
    - (.*?)/head/(.*?)
  rateLimitPeriod: 100  // time in seconds
  rateLimitCount: 100
  cacheMaxItems: 2000   // max size of LRU Cache
```  