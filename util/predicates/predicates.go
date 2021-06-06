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

package predicates

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func ControlledIngressPredicate(ctx context.Context, crclient client.Client, controlller string) predicate.Funcs {
	log := log.FromContext(ctx)
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		ing, ok := obj.(*networkingv1.Ingress)
		if !ok {
			err := fmt.Errorf("expected a Ingress but got a %T", obj)
			log.Error(err, "unable to find controlled ingress")
			return false
		}

		if ing.Spec.IngressClassName == nil {
			return false
		}

		ic := &networkingv1.IngressClass{}
		if err := crclient.Get(ctx, client.ObjectKey{Name: *ing.Spec.IngressClassName}, ic); err != nil {
			return false
		}

		return ic.Spec.Controller == controlller
	})
}
