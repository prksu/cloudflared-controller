module github.com/prksu/cloudflared-controller

go 1.16

require (
	github.com/cloudflare/cloudflared v0.0.0-20210527163216-98a0844f5619
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.2.0
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.1.2
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	go.uber.org/multierr v1.5.0
	go.uber.org/zap v1.15.0
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/yaml v1.2.0
)
