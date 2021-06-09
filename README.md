# Cloudflared Controller

Cloudflared Controller is a Kubernetes controller that turns Cloudflare tunnel (formerly known as Argo tunnel) as Kubernetes LoadBalancerIngress.

## Project Status

This project is currently a work-in-progress which only has initial implementation and some lack of features. For more details see our [roadmap](#roadmap) and the [issue trakcer on Github](https://github.com/prksu/cloudflared-controller/issues)

## Installation

Check out the [installation guide](./docs/installation.md)

## Getting Started

Check out the [getting started guide](./docs/getting-started.md)

## Roadmap

- Support cloudflare tunnel [lb route](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/routing-to-tunnel/lb) type
- Support Kubernetes Service [LoadBalancerClass](https://kubernetes.io/docs/concepts/services-networking/service/#load-balancer-class) (Kubernetes v1.21)
