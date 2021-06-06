/*
Copyright 2021 Ahmad Nurus S.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/prksu/cloudflared-controller/util"
	"github.com/prksu/cloudflared-controller/util/patch"
	"github.com/prksu/cloudflared-controller/util/predicates"
	"github.com/prksu/cloudflared-controller/util/resources"
)

const (
	IngressControllerName = "cloudflared.cloudflare.com/ingress-controller"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingressclasses,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update
// +kubebuilder:rbac:groups=cloudflared.cloudflare.com,resources=tunnels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflared.cloudflare.com,resources=tunnelconfigurations,verbs=get;list;watch

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}, builder.WithPredicates(predicates.ControlledIngressPredicate(ctx, r.Client, IngressControllerName))).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := log.FromContext(ctx)
	ing := &networkingv1.Ingress{}
	if err := r.Client.Get(ctx, req.NamespacedName, ing); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Ingress resource not found or already deleted")
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to fetch Ingress resource")
		return ctrl.Result{}, err
	}

	patcher, err := patch.NewPatcher(ing, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if err := patcher.Patch(ctx, ing); err != nil {
			log.Error(err, "unable patch Ingress resource")
			reterr = err
		}
	}()

	ir := resources.NewIngressResources(ing)
	tunnel := ir.Tunnel()
	tunnelConfig, err := util.TunnelConfigurationFromIngress(ctx, r.Client, ing)
	if err != nil {
		return ctrl.Result{}, err
	}

	if _, err := controllerutil.CreateOrPatch(ctx, r.Client, tunnel, func() error {
		tunnel.Spec.TunnelConfigurationSpec = tunnelConfig.Spec
		tunnel.Spec.IngressRules = ir.TunnelIngressRules()
		return controllerutil.SetControllerReference(ing, tunnel, r.Scheme)
	}); err != nil {
		return ctrl.Result{}, err
	}

	if tunnel.Status.Zone == "" {
		return ctrl.Result{Requeue: true}, nil
	}

	ing.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{
		{
			Hostname: tunnel.Status.Zone,
		},
	}

	return ctrl.Result{}, nil
}
