/*
Copyright 2021 The Kubernetes Authors.

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

package patch

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPatcher_Patch(t *testing.T) {
	tests := []struct {
		name    string
		obj     client.Object
		f       func(client.Client, client.Object) error
		wantErr bool
	}{
		{
			name: "patch pod ownerref",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "patch-pod-ownerref",
					Namespace: "default",
				},
			},
			f: func(c client.Client, o client.Object) error {
				ctx := context.Background()
				obj := o.(*corev1.Pod)
				if err := c.Create(ctx, obj); err != nil {
					return err
				}

				p, err := NewPatcher(obj, c)
				if err != nil {
					return err
				}

				refs := []metav1.OwnerReference{
					{
						APIVersion: "apps/v1",
						Kind:       "ReplicaSet",
						Name:       "fake-rs",
						UID:        types.UID("fake-uid"),
					},
				}
				obj.SetOwnerReferences(refs)
				if err := p.Patch(ctx, obj); err != nil {
					return err
				}

				afterObj := obj.DeepCopy()
				if err := c.Get(ctx, client.ObjectKeyFromObject(obj), o); err != nil {
					return err
				}

				if !reflect.DeepEqual(obj.OwnerReferences, afterObj.OwnerReferences) {
					return fmt.Errorf("expected = %v, got = %v", obj.OwnerReferences, afterObj.OwnerReferences)
				}

				return nil
			},
		},
		{
			name: "patch pod spec",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "patch-pod-spec",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "bar",
				},
			},
			f: func(c client.Client, o client.Object) error {
				ctx := context.Background()
				obj := o.(*corev1.Pod)
				if err := c.Create(ctx, obj); err != nil {
					return err
				}

				p, err := NewPatcher(obj, c)
				if err != nil {
					return err
				}

				obj.Spec.ServiceAccountName = "baz"
				if err := p.Patch(ctx, obj); err != nil {
					return err
				}

				afterObj := obj.DeepCopy()
				if err := c.Get(ctx, client.ObjectKeyFromObject(obj), o); err != nil {
					return err
				}

				if !reflect.DeepEqual(obj.Spec, afterObj.Spec) {
					return fmt.Errorf("expected = %v, got = %v", obj.Spec, afterObj.Spec)
				}

				return nil
			},
		},
		{
			name: "patch pod status",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "patch-pod-status",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:               corev1.PodInitialized,
							Status:             corev1.ConditionTrue,
							LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, metav1.Now().Location()),
						},
					},
				},
			},
			f: func(c client.Client, o client.Object) error {
				ctx := context.Background()
				obj := o.(*corev1.Pod)
				if err := c.Create(ctx, obj); err != nil {
					return err
				}

				p, err := NewPatcher(obj, c)
				if err != nil {
					return err
				}

				obj.Status.Conditions = append(obj.Status.Conditions, corev1.PodCondition{
					Type:               corev1.ContainersReady,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.Date(2015, 1, 1, 12, 0, 0, 0, metav1.Now().Location()),
					Reason:             "reason",
					Message:            "message",
				})
				if err := p.Patch(ctx, obj); err != nil {
					return err
				}

				afterObj := obj.DeepCopy()
				if err := c.Get(ctx, client.ObjectKeyFromObject(obj), o); err != nil {
					return err
				}

				if !reflect.DeepEqual(obj.Spec, afterObj.Spec) {
					return fmt.Errorf("expected = %v, got = %v", obj.Spec, afterObj.Spec)
				}

				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := fake.NewClientBuilder().Build()
			if err := tt.f(fc, tt.obj); (err != nil) != tt.wantErr {
				t.Errorf("Patcher.Patch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
