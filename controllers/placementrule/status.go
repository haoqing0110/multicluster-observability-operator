// Copyright (c) 2021 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package placementrule

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	mcov1beta1 "github.com/open-cluster-management/multicluster-observability-operator/api/v1beta1"
	"github.com/open-cluster-management/multicluster-observability-operator/pkg/util"
)

var (
	statusMap = map[string]string{
		"Available":    "Available",
		"Progressing":  "Progressing",
		"Deployed":     "Progressing",
		"Disabled":     "Degraded",
		"Degraded":     "Degraded",
		"NotSupported": "Degraded",
	}
)

func updateAddonStatus(c client.Client, addonList mcov1beta1.ObservabilityAddonList) error {
	for _, addon := range addonList.Items {
		if addon.Status.Conditions == nil || len(addon.Status.Conditions) == 0 {
			continue
		}
		conditions := []metav1.Condition{}
		for _, c := range addon.Status.Conditions {
			condition := metav1.Condition{
				Type:               statusMap[c.Type],
				Status:             c.Status,
				LastTransitionTime: c.LastTransitionTime,
				Reason:             c.Reason,
				Message:            c.Message,
			}
			conditions = append(conditions, condition)
		}
		managedclusteraddon := &addonv1alpha1.ManagedClusterAddOn{}
		err := c.Get(context.TODO(), types.NamespacedName{
			Name:      util.ManagedClusterAddonName,
			Namespace: addon.ObjectMeta.Namespace,
		}, managedclusteraddon)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Info("managedclusteraddon does not exist", "namespace", addon.ObjectMeta.Namespace)
				continue
			}
			log.Error(err, "Failed to get managedclusteraddon", "namespace", addon.ObjectMeta.Namespace)
			return err
		}
		if !reflect.DeepEqual(conditions, managedclusteraddon.Status.Conditions) {
			managedclusteraddon.Status.Conditions = conditions
			err = c.Status().Update(context.TODO(), managedclusteraddon)
			if err != nil {
				log.Error(err, "Failed to update status for managedclusteraddon", "namespace", addon.ObjectMeta.Namespace)
				return err
			}
			log.Info("Updated status for managedclusteraddon", "namespace", addon.ObjectMeta.Namespace)
		}
	}
	return nil
}
