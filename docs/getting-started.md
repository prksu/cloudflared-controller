# Getting Started

We assume Cloudflared Controller already installed in your cluster. Otherwise see [installation guide](./installation.md)

## Prerequisites

- Install [cloudflared](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation) command line tool

## Getting Started with Ingress

We assume you have service called *my-nginx* in the current namespace. Otherwise you can following [this example](https://kubernetes.io/docs/concepts/services-networking/connect-applications-service/) to deploy sample service

### Create OriginCert Secret

Obtains cloudflare tunnel origincert by login using cloudflared command line tool

```bash
cloudflared login
```

Create Secret for tunnel origincert

```bash
kubectl create secret generic default-origincert --from-file=cert.pem=$HOME/.cloudflared/cert.pem
```

### Create TunnelConfiguration

Create TunnelConfiguration for IngressClass reference parameters as following

```yaml
apiVersion: cloudflared.cloudflare.com/v1alpha1
kind: TunnelConfiguration
metadata:
  name: tunnelconfiguration-sample
spec:
  originCert:
    kind: Secret
    name: default-origincert
```

Or by using the examples

```bash
kubectl apply -f docs/example/tunnelconfiguration.yaml
```

### Create an Ingress

Create an IngressClass as following

```yaml
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: cloudflared
spec:
  controller: cloudflared.cloudflare.com/ingress-controller
  parameters:
    apiGroup: cloudflared.cloudflare.com
    kind: TunnelConfiguration
    name: tunnelconfiguration-sample
```

Or by using the examples

```bash
kubectl apply -f docs/example/ingressclass.yaml
```

Create an Ingress with specific ingressClassName

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cloudflared-ingress
spec:
  ingressClassName: cloudflared
  defaultBackend:
    service:
      name: my-nginx
      port:
        number: 80
```
