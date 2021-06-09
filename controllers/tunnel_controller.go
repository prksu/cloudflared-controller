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
	"encoding/json"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudflaredv1alpha1 "github.com/prksu/cloudflared-controller/api/v1alpha1"
	"github.com/prksu/cloudflared-controller/cloudflare"
	"github.com/prksu/cloudflared-controller/util"
	"github.com/prksu/cloudflared-controller/util/patch"
	"github.com/prksu/cloudflared-controller/util/resources"
	"github.com/prksu/cloudflared-controller/util/tunnelroutes"
)

const (
	TunnelControllerName = "cloudflared.cloudflare.com/tunnel-controller"
)

// TunnelReconciler reconciles a Tunnel object
type TunnelReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=cloudflared.cloudflare.com,resources=tunnels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloudflared.cloudflare.com,resources=tunnels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloudflared.cloudflare.com,resources=tunnels/finalizers,verbs=update

// SetupWithManager sets up the controller with the Manager.
func (r *TunnelReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudflaredv1alpha1.Tunnel{}).
		Complete(r)
}

func (r *TunnelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := log.FromContext(ctx)
	tunnel := &cloudflaredv1alpha1.Tunnel{}
	if err := r.Client.Get(ctx, req.NamespacedName, tunnel); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Tunnel resource not found or already deleted")
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to fetch Tunnel resource")
		return ctrl.Result{}, err
	}

	ocsecret, err := util.GetOriginCertSecret(ctx, r.Client, req.Namespace, tunnel.Spec.OriginCert)
	if err != nil {
		return ctrl.Result{}, err
	}

	cfclient, err := cloudflare.NewClient(
		cloudflare.WithAPIToken(os.Getenv(cloudflare.APITokenEnv)),
		cloudflare.WithOriginCert(ocsecret.Data["cert.pem"]),
		cloudflare.WithLogger(log),
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	patcher, err := patch.NewPatcher(tunnel, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if err := patcher.Patch(ctx, tunnel); err != nil {
			log.Error(err, "unable patch Tunnel resource")
			reterr = err
		}
	}()

	if !tunnel.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, cfclient, tunnel)
	}

	return r.reconcile(ctx, cfclient, tunnel)
}

func (r *TunnelReconciler) reconcile(ctx context.Context, cfclient cloudflare.Client, tunnel *cloudflaredv1alpha1.Tunnel) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	controllerutil.AddFinalizer(tunnel, cloudflaredv1alpha1.TunnelFinalizer)
	log.Info("Reconciling")

	tr := resources.NewTunnelResources(tunnel)
	cftunnelName := tr.TunnelName()

	log.Info("Ensuring cloudflare zone")
	cfzone, err := cfclient.Zones().Get(ctx, cfclient.ZoneID())
	if err != nil {
		return ctrl.Result{}, err
	}

	tunnel.Status.Zone = cfzone.Name

	log.Info("Ensuring cloudflare tunnel")
	cftunnel, err := cfclient.Tunnels().GetByName(ctx, cftunnelName)
	if err != nil {
		if err != cloudflare.ErrNotFound {
			return ctrl.Result{}, err
		}

		log.Info("Creating new cloudflare tunnel")
		cftunnel, err = cfclient.Tunnels().Create(ctx, cftunnelName)
		if err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Creating tunnel secret")
		secretData := make(map[string][]byte)
		secretDataKey := cftunnelName + ".json"
		secretDataValue, err := json.Marshal(cftunnel.CredentialsFile)
		if err != nil {
			// should we delete the cloudflare tunnel?
			return ctrl.Result{}, err
		}

		secretData[secretDataKey] = secretDataValue
		secret := tr.Secret(secretData)
		if err := controllerutil.SetControllerReference(tunnel, secret, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Client.Create(ctx, secret); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Info("Ensuring cloudflare tunnel route")
	desiredRoutes := tunnelroutes.FromTunnelSpec(tunnel.Spec)
	// If we got empty routes from TunnelSpec
	// then use the zone name (root domain).
	if len(desiredRoutes) == 0 {
		desiredRoutes = append(desiredRoutes, cfzone.Name)
	}

	actualRoutes := tunnelroutes.FromTunnelStatus(tunnel.Status)
	hostnameNeedRouted := tunnelroutes.Difference(desiredRoutes, actualRoutes)
	for _, hostname := range hostnameNeedRouted {
		log.Info("Updating tunnel route")
		cftunnelRoute := &cloudflare.TunnelDNSRoute{
			Hostname: hostname,
		}

		if err := cfclient.Tunnels().Route(ctx, cftunnel.ID, cftunnelRoute); err != nil {
			return ctrl.Result{}, err
		}
	}
	tunnel.Status.Routes = desiredRoutes

	log.Info("Ensuring tunnel secret")
	secret := &corev1.Secret{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: tunnel.Namespace, Name: tr.SecretName()}, secret); err != nil {
		return ctrl.Result{}, err
	}

	configMap := tr.ConfigMap()
	configMapOp, err := controllerutil.CreateOrPatch(ctx, r.Client, configMap, func() error {
		configMap.Data, err = tr.ConfigMapData()
		if err != nil {
			return err
		}

		return controllerutil.SetControllerReference(tunnel, configMap, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Reconcile tunnel configmap", "operation", configMapOp)

	dep := tr.Deployment()
	depOp, err := controllerutil.CreateOrPatch(ctx, r.Client, dep, func() error {
		// Restart the cloudflared daemon if the configuration got an update by set an annotations
		// into PodTemplateSpec to gracefully restart the Pod. So the Tunnel could be use new configurations.
		// This restart mechanism inspired by kubectl rollout restart
		if configMapOp == controllerutil.OperationResultUpdated {
			dep.Spec.Template.SetAnnotations(map[string]string{
				"cloudflared.cloudflare.com/restarted-at": time.Now().Format(time.RFC3339),
			})
		}

		dep.Spec.Replicas = pointer.Int32Ptr(1)
		return controllerutil.SetControllerReference(tunnel, dep, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Reconcile tunnel deployment", "operation", depOp)
	return ctrl.Result{}, nil
}

func (r *TunnelReconciler) reconcileDelete(ctx context.Context, cfclient cloudflare.Client, tunnel *cloudflaredv1alpha1.Tunnel) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling to deleting")

	tr := resources.NewTunnelResources(tunnel)
	cftunnelName := tr.TunnelName()

	log.Info("Ensuring all tunnel daemon has stopped")
	dep := tr.Deployment()
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(dep), dep); err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if *dep.Spec.Replicas != 0 {
		log.Info("Scaling tunnel daemon to 0")
		depPatch := client.MergeFrom(dep.DeepCopyObject())
		dep.Spec.Replicas = pointer.Int32Ptr(0)
		// Stopping the cloudflared tunnel daemon by scaling dep to 0
		// so cloudflared tunnel will be gracefully shutting down and hopefully closing all connections.
		if err := r.Client.Patch(ctx, dep, depPatch); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	log.Info("Deleting cloudflare tunnel")
	cftunnel, err := cfclient.Tunnels().GetByName(ctx, cftunnelName)
	if err != nil && err != cloudflare.ErrNotFound {
		return ctrl.Result{}, err
	}

	if err := cfclient.Tunnels().Delete(ctx, cftunnel.ID); err != nil && err != cloudflare.ErrNotFound {
		return ctrl.Result{}, err
	}

	controllerutil.RemoveFinalizer(tunnel, cloudflaredv1alpha1.TunnelFinalizer)
	return ctrl.Result{}, nil
}
