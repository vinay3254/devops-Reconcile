/*
Copyright 2026.

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

package controller

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	monitoringv1alpha1 "github.com/statuspage/operator/api/v1alpha1"
)

// MonitorReconciler reconciles a Monitor object
type MonitorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=monitoring.statuspage.io,resources=monitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.statuspage.io,resources=monitors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=monitoring.statuspage.io,resources=monitors/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is the core reconciliation loop. It is called whenever a Monitor
// resource is created, updated, or deleted — and also when owned CronJobs change.
func (r *MonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Fetch the Monitor resource
	monitor := &monitoringv1alpha1.Monitor{}
	if err := r.Get(ctx, req.NamespacedName, monitor); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted — owned CronJob will be garbage collected via OwnerReference
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling Monitor", "name", monitor.Name, "url", monitor.Spec.URL)

	// 2. Reconcile the CronJob for this Monitor
	if err := r.reconcileCronJob(ctx, monitor); err != nil {
		logger.Error(err, "Failed to reconcile CronJob")
		return ctrl.Result{}, err
	}

	// 3. Read the latest check result from the ConfigMap written by the check job
	if err := r.syncStatusFromConfigMap(ctx, monitor); err != nil {
		// Non-fatal: ConfigMap may not exist yet (no checks have run)
		logger.V(1).Info("No check results yet", "reason", err.Error())
	}

	return ctrl.Result{}, nil
}

// reconcileCronJob ensures a CronJob exists and matches the Monitor spec.
func (r *MonitorReconciler) reconcileCronJob(ctx context.Context, monitor *monitoringv1alpha1.Monitor) error {
	logger := log.FromContext(ctx)
	cronJobName := cronJobNameFor(monitor)

	// Convert intervalSeconds to a cron schedule
	schedule := intervalToCron(monitor.Spec.IntervalSeconds)

	desired := r.buildCronJob(monitor, cronJobName, schedule)

	// Check if CronJob already exists
	existing := &batchv1.CronJob{}
	err := r.Get(ctx, types.NamespacedName{Name: cronJobName, Namespace: monitor.Namespace}, existing)

	if errors.IsNotFound(err) {
		// Create it
		logger.Info("Creating CronJob", "cronJob", cronJobName)
		if err := ctrl.SetControllerReference(monitor, desired, r.Scheme); err != nil {
			return err
		}
		if err := r.Create(ctx, desired); err != nil {
			return err
		}
		// Update status with job name
		monitor.Status.CheckJobName = cronJobName
		return r.Status().Update(ctx, monitor)
	}
	if err != nil {
		return err
	}

	// CronJob exists — update if schedule or image changed
	if existing.Spec.Schedule != schedule {
		logger.Info("Updating CronJob schedule", "cronJob", cronJobName, "schedule", schedule)
		existing.Spec.Schedule = schedule
		existing.Spec.JobTemplate = desired.Spec.JobTemplate
		return r.Update(ctx, existing)
	}

	return nil
}

// buildCronJob constructs the CronJob that runs HTTP checks for a Monitor.
func (r *MonitorReconciler) buildCronJob(monitor *monitoringv1alpha1.Monitor, name, schedule string) *batchv1.CronJob {
	configMapName := configMapNameFor(monitor)
	timeout := monitor.Spec.TimeoutSeconds
	if timeout == 0 {
		timeout = 10
	}

	// The check container uses curl to hit the URL and writes results to a ConfigMap
	// via kubectl. In Phase 1 we keep this simple; Phase 2 will write to Postgres.
	checkScript := fmt.Sprintf(`
set -e
URL="%s"
EXPECTED=%d
TIMEOUT=%d
NAMESPACE="%s"
CM_NAME="%s"

START=$(date +%%s%%3N)
HTTP_CODE=$(curl -s -o /dev/null -w "%%{http_code}" --max-time $TIMEOUT "$URL" 2>/dev/null || echo "000")
END=$(date +%%s%%3N)
LATENCY=$((END - START))

if [ "$HTTP_CODE" = "$EXPECTED" ]; then
    RESULT="UP"
else
    RESULT="DOWN"
fi

TIMESTAMP=$(date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ)

kubectl create configmap "$CM_NAME" \
    --from-literal=lastResult="$RESULT" \
    --from-literal=lastStatusCode="$HTTP_CODE" \
    --from-literal=lastCheck="$TIMESTAMP" \
    --from-literal=latencyMs="$LATENCY" \
    -n "$NAMESPACE" \
    --dry-run=client -o yaml | kubectl apply -f -

echo "Check complete: URL=$URL result=$RESULT code=$HTTP_CODE latency=${LATENCY}ms"
`, monitor.Spec.URL, monitor.Spec.ExpectedStatus, timeout, monitor.Namespace, configMapName)

	ttlSeconds := int32(3600) // keep completed jobs for 1h
	failedJobsLimit := int32(3)
	successfulJobsLimit := int32(3)
	backoffLimit := int32(0) // don't retry — just wait for next schedule

	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: monitor.Namespace,
			Labels:    labelsFor(monitor),
		},
		Spec: batchv1.CronJobSpec{
			Schedule:                   schedule,
			ConcurrencyPolicy:          batchv1.ForbidConcurrent,
			FailedJobsHistoryLimit:     &failedJobsLimit,
			SuccessfulJobsHistoryLimit: &successfulJobsLimit,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					BackoffLimit:            &backoffLimit,
					TTLSecondsAfterFinished: &ttlSeconds,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: labelsFor(monitor),
						},
						Spec: corev1.PodSpec{
							RestartPolicy:      corev1.RestartPolicyNever,
							ServiceAccountName: "monitor-checker",
							Containers: []corev1.Container{
								{
									Name:  "checker",
									Image: "bitnami/kubectl:latest",
									Command: []string{
										"/bin/sh",
										"-c",
										checkScript,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// syncStatusFromConfigMap reads the check result ConfigMap and updates Monitor.Status.
func (r *MonitorReconciler) syncStatusFromConfigMap(ctx context.Context, monitor *monitoringv1alpha1.Monitor) error {
	cm := &corev1.ConfigMap{}
	cmName := configMapNameFor(monitor)
	if err := r.Get(ctx, types.NamespacedName{Name: cmName, Namespace: monitor.Namespace}, cm); err != nil {
		return err
	}

	// Parse results from ConfigMap
	result := monitoringv1alpha1.MonitorResult(cm.Data["lastResult"])
	if result == "" {
		result = monitoringv1alpha1.MonitorResultUnknown
	}

	var statusCode int32
	fmt.Sscanf(cm.Data["lastStatusCode"], "%d", &statusCode)

	now := metav1.Now()
	monitor.Status.LastCheck = &now
	monitor.Status.LastResult = result
	monitor.Status.LastStatusCode = statusCode

	// Set condition
	conditionStatus := metav1.ConditionTrue
	conditionReason := "CheckPassed"
	conditionMsg := fmt.Sprintf("HTTP %d from %s", statusCode, monitor.Spec.URL)
	if result != monitoringv1alpha1.MonitorResultUp {
		conditionStatus = metav1.ConditionFalse
		conditionReason = "CheckFailed"
		conditionMsg = fmt.Sprintf("Got HTTP %d, expected %d from %s", statusCode, monitor.Spec.ExpectedStatus, monitor.Spec.URL)
	}

	meta := metav1.Condition{
		Type:               "Healthy",
		Status:             conditionStatus,
		ObservedGeneration: monitor.Generation,
		LastTransitionTime: now,
		Reason:             conditionReason,
		Message:            conditionMsg,
	}
	setCondition(&monitor.Status.Conditions, meta)

	return r.Status().Update(ctx, monitor)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.Monitor{}).
		Owns(&batchv1.CronJob{}).
		Named("monitor").
		Complete(r)
}

// --- Helpers ---

func cronJobNameFor(m *monitoringv1alpha1.Monitor) string {
	return fmt.Sprintf("monitor-%s", m.Name)
}

func configMapNameFor(m *monitoringv1alpha1.Monitor) string {
	return fmt.Sprintf("monitor-result-%s", m.Name)
}

func labelsFor(m *monitoringv1alpha1.Monitor) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":           "statuspage-operator",
		"app.kubernetes.io/component":      "monitor-checker",
		"monitoring.statuspage.io/monitor": m.Name,
	}
}

// intervalToCron converts intervalSeconds to the closest valid cron expression.
// Cron minimum granularity is 1 minute.
func intervalToCron(intervalSeconds int32) string {
	minutes := intervalSeconds / 60
	if minutes < 1 {
		minutes = 1
	}
	if minutes == 1 {
		return "* * * * *"
	}
	return fmt.Sprintf("*/%d * * * *", minutes)
}

// setCondition upserts a condition into the conditions slice.
func setCondition(conditions *[]metav1.Condition, newCond metav1.Condition) {
	for i, c := range *conditions {
		if c.Type == newCond.Type {
			(*conditions)[i] = newCond
			return
		}
	}
	*conditions = append(*conditions, newCond)
}
