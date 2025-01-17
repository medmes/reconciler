package service

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/reconciler/pkg/reconciler"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"
	"time"

	k8s "github.com/kyma-incubator/reconciler/pkg/reconciler/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createOrUpdateStatusCm(ctx context.Context, task *reconciler.Task, status reconciler.Status, kubeclient k8s.Client, logger *zap.SugaredLogger) error {
	configMapName := fmt.Sprintf("%s-status", strings.ToLower(task.Component))
	if task.Namespace == "" {
		task.Namespace = "default"
	}
	clientset, err := kubeclient.Clientset()
	if err != nil {
		logger.Errorf("Error getting clientset: %s", err)
		return err
	}
	_, err = clientset.CoreV1().Namespaces().Get(ctx, task.Namespace, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		logger.Debugf("Namespace %s not found, status ConfigMap will only be created if namespace exists", task.Namespace)
		return nil
	}
	// Get ConfigMap
	configMap, err := clientset.CoreV1().ConfigMaps(task.Namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		// ConfigMap does not exist, create new one
		logger.Debugf("ConfigMap '%s' not found, creating new one", configMapName)
		lastReconciliation := ""
		if status != reconciler.StatusRunning && status != reconciler.StatusNotstarted {
			lastReconciliation = time.Now().String()
		}
		configFile := map[string]string{
			"name":                task.Component,
			"version":             task.Version,
			"status":              string(status),
			"last-reconciliation": lastReconciliation,
		}
		cm := corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapName,
				Namespace: task.Namespace,
				Labels: map[string]string{
					"reconciler.kyma-project.io/managed-by":             "reconciler",
					"reconciler.kyma-project.io/origin-version":         task.Version,
					"reconciler.kyma-project.io/reconciliation-summary": "true",
				},
			},
			Data: configFile,
		}
		_, err := clientset.CoreV1().ConfigMaps(task.Namespace).Create(ctx, &cm, metav1.CreateOptions{})
		if err != nil {
			logger.Debugf("Error after creating ConfigMap '%s': %s", configMapName, err)
			return err
		}

	} else if err != nil {
		// Error while fetching ConfigMap
		logger.Errorf("Error getting ConfigMap '%s': %s", configMapName, err)
		return err
	} else {
		// Update existing ConfigMap
		logger.Debugf("ConfigMap '%s' found, updating status", configMapName)
		configMap.Data["version"] = task.Version
		configMap.Data["status"] = string(status)
		if status != reconciler.StatusRunning && status != reconciler.StatusNotstarted {
			configMap.Data["last-reconciliation"] = time.Now().String()
		}
		_, err := clientset.CoreV1().ConfigMaps(task.Namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		if err != nil {
			logger.Debugf("Error updating ConfigMap '%s': %s", configMapName, err)
			return err
		}
		logger.Debugf("ConfigMap '%s' successfully updated", configMapName)
	}
	return nil
}
