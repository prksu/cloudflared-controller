# Cloudflared Controller Installation

Currently, we do not have any release yet. So for installation, you need to clone this repository into your local and build your own controller image. We assume every command below invoked under the project directory.

## Prerequisites

- A Kubernetes v1.18+
- A Cloudflare account with registered zone
- A Cloudflare API Token with DNS Zone Edit permission. A [reference](https://developers.cloudflare.com/api/tokens/create) how to create API Token

## Setup required environment variable

```bash
export CF_API_TOKEN=<token>
export CF_B64API_TOKEN=$(echo -n $CF_API_TOKEN | base64 | tr -d '\n')
```

## Build and Push the controller image

```bash
export IMG=<your-registry>/cloudflared-controller 
make docker-build 
make docker-push
```

## Install the CRD and deploy the controller

```bash
make install && make deploy
```
