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
  host: tezos      // should be `tezos`for recommended Setup
  port: 8732       // Tezos Node Port

server:
  port: 80       // Port that TezProxy will run at

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
    - /chains/main/blocks$
  rateLimitPeriod: 100  // time in seconds
  rateLimitCount: 100
  cacheMaxItems: 2000   // max size of LRU Cache
```  

## Docker

Autmated Docker Builds for TezProxy are available at:

https://hub.docker.com/r/tezexinfo/tezproxy/

## Recommended Setup

If you plan to run TezProxy as a public api endpoint ( for apps like tezbox ) we recommend the following Setup:


1.) Start with any number of Servers, preferably in different Datacenters

2.) install Docker and docker-compose on all Servers: https://docs.docker.com/install/

3.) Pick one Server as your Swarm Manager and run `docker swarm init`

4.) join all other Servers into the swarm using the command in the output of #3

5.) On the Swarm Manager: Copy the confiy.yml - Example from above and import it into Docker Swarm Configs

```
$ docker config create tezproxy_config config.yml
```

6.) Run the following Commands to install the Tezos Node and TezRPC

```
$ wget https://raw.githubusercontent.com/tezexInfo/TezProxy/master/swarmCompose/docker-compose.yml
$ docker stack deploy tezproxy -c docker-compose.yml
```

After this it might take your node some time to generate a new identity and sync the blockchain, but after a few minutes
your TezProxy shouldbe ready to go!


If you want yo can also install a very simple Monitoring Service that will check your Tezos Nodes, TezProxy instances
as well as your general Server Statistics by following the guide on:

https://github.com/stefanprodan/swarmprom

And after that, you will get a Monitoring Dashboard that will look a lot like:

![monitoring](https://raw.githubusercontent.com/tezexInfo/TezProxy/master/tezproxy_mon.png "TezProxy Moitoring")
