/*
Copyright 2025.

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
	"strings"
	"time"

	clients "github.com/konflux-ci/kite/bridge-operator/internal/clients"
	"github.com/sirupsen/logrus"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PipelineRunReconciler reconciles a PipelineRun object
type PipelineRunReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	KiteClient *clients.KiteClient
	Logger     *logrus.Logger
}

// +kubebuilder:rbac:groups=tekton.konflux.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tekton.konflux.dev,resources=pipelineruns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tekton.konflux.dev,resources=pipelineruns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// Here we compare the state specified by the PipelineRun object
// against the actual cluster state, and then create or resolve pipeline run issue records
// via the KITE service.
func (r *PipelineRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the PipelineRun instance by namespace and name
	var pipelineRun v1.PipelineRun
	if err := r.Get(ctx, req.NamespacedName, &pipelineRun); err != nil {
		// In the Reconcile path the only expected error on a Get is "NotFound".
		// In this case the Pipeline was deleted, so do nothing.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Lets only process completed PipelineRuns
	if pipelineRun.Status.CompletionTime == nil {
		r.Logger.WithFields(logrus.Fields{
			"pipeline_run": pipelineRun.Name,
			"namespace":    pipelineRun.Namespace,
		}).Debug("PipelineRun not yet completed, skipping")
		return ctrl.Result{}, nil
	}

	// Determine status of PipelineRun
	status := r.getPipelineRunStatus(&pipelineRun)

	logFields := logrus.Fields{
		"pipeline_run": pipelineRun.Name,
		"namespace":    pipelineRun.Namespace,
		"status":       status,
	}
	logEntry := r.Logger.WithFields(logFields)

	// Handle PipelineRun based on status
	switch status {
	case "failed":
		logEntry.Info("Processing failed PipelineRun")
		return r.handlePipelineRunFailure(ctx, &pipelineRun)
	case "succeeded":
		logEntry.Info("Processing successful PipelineRun")
		return r.handlePipelineRunSuccess(ctx, &pipelineRun)
	default:
		logEntry.Debugf("Ignoring PipelineRun with status: %s", status)
		return ctrl.Result{}, nil
	}
}

func (r *PipelineRunReconciler) handlePipelineRunFailure(ctx context.Context, pr *v1.PipelineRun) (ctrl.Result, error) {
	failureReason := r.getFailureReason(ctx, pr)
	pipelineName := r.getPipelineName(pr)

	// Payload sent to KITE (/api/v1/webhooks/pipeline-failure)
	payload := clients.PipelineFailurePayload{
		PipelineName:  pipelineName,
		Namespace:     pr.Namespace,
		FailureReason: failureReason,
		RunID:         string(pr.UID),
		Severity:      r.determineSeverity(pr),
	}

	// In the event of failure, retry in x minutes
	if err := r.KiteClient.ReportPipelineFailure(payload); err != nil {
		r.Logger.WithError(err).WithFields(logrus.Fields{
			"id":           pr.UID,
			"pipeline_run": pr.Name,
			"namespace":    pr.Namespace,
			"operation":    "pipeline-failure",
		}).Error("An error occurred when reporting a pipeline failure from controller.")

		// Try again in 2 minutes...
		return ctrl.Result{RequeueAfter: time.Minute * 2}, fmt.Errorf("failed to report pipeline failure from controller")
	}

	r.Logger.WithFields(logrus.Fields{
		"pipeline_run": pr.Name,
		"id":           pr.UID,
		"operation":    "pipeline-failure",
	}).Info("Successfully reported pipeline failure to KITE")

	return ctrl.Result{}, nil
}

func (r *PipelineRunReconciler) handlePipelineRunSuccess(ctx context.Context, pr *v1.PipelineRun) (ctrl.Result, error) {
	pipelineName := r.getPipelineName(pr)
	// Payload sent to KITE (/api/v1/webhooks/pipeline-success)
	payload := clients.PipelineSuccessPayload{
		PipelineName: pipelineName,
		Namespace:    pr.Namespace,
	}

	// In the event of failure, retry in x minutes
	if err := r.KiteClient.ReportPipelineSuccess(payload); err != nil {
		r.Logger.WithError(err).WithFields(logrus.Fields{
			"id":           pr.UID,
			"pipeline_run": pr.Name,
			"namespace":    pr.Namespace,
			"operation":    "pipeline-success",
		}).Error("An error occurred when reporting a successful pipeline from controller.")
		// Retry in 2 minutes...
		return ctrl.Result{RequeueAfter: time.Minute * 2}, fmt.Errorf("failed to report pipeline success from controller")
	}

	r.Logger.WithFields(logrus.Fields{
		"pipeline_run": pr.Name,
		"id":           pr.UID,
		"operation":    "pipeline-success",
	}).Info("Successfully reported pipeline success to KITE")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		// For().
		Named("pipelinerun").
		Complete(r)
}

// getPipelineRunStatus returns the status of the PipelineRun by checking
// the type and status of each condition in the PipelineRun status.
func (p *PipelineRunReconciler) getPipelineRunStatus(pr *v1.PipelineRun) string {
	if pr.Status.Conditions == nil {
		return "unknown"
	}

	for _, condition := range pr.Status.Conditions {
		// Only check completed conditions
		if condition.Type == "Succeeded" {
			switch condition.Status {
			case "True":
				return "succeeded"
			case "False":
				return "failed"
			}
		}
	}

	p.Logger.WithFields(logrus.Fields{
		"pipeline_run": pr.Name,
	}).Debug("Could not determine PipelineRun status.")

	return "unknown"
}

func (p *PipelineRunReconciler) getPipelineName(pr *v1.PipelineRun) string {
	// Check Spec first
	if pr.Spec.PipelineRef != nil && pr.Spec.PipelineRef.Name != "" {
		return pr.Spec.PipelineRef.Name
	}

	// Next check standard label
	if pipelineName, exists := pr.Labels["tekton.dev/pipeline"]; exists {
		return pipelineName
	}

	// Fallback to PipelineRun name
	p.Logger.WithFields(logrus.Fields{
		"pipeline_run": pr.Name,
	}).Debug("Unable to extract Pipeline name, falling back to PipelineRun name")

	return pr.Name
}

func (r *PipelineRunReconciler) getFailureReason(ctx context.Context, pr *v1.PipelineRun) string {
	// First, lets look for PipelineRun failure reasons in the conditions
	if pr.Status.Conditions != nil {
		for _, condition := range pr.Status.Conditions {
			// If PipelineRun ran and failed...
			if condition.Type == "Succeeded" && condition.Status == "False" {
				// Check reasons for failure
				if condition.Message != "" {
					return condition.Message
				}
				if condition.Reason != "" {
					return condition.Reason
				}
			}
		}
	}

	// Next, lets look at child references (specifically TaskRuns) for failure reasons
	if pr.Status.ChildReferences != nil {
		failedTasks := r.getFailedTasksFromChildReferences(ctx, pr)
		if len(failedTasks) > 0 {
			return fmt.Sprintf("Failed pipeline tasks: %s", strings.Join(failedTasks, ", "))
		}
	}

	r.Logger.WithFields(logrus.Fields{
		"pipeline_run": pr.Name,
	}).Debug("Could not determine reason for failure.")

	return "PipelineRun failed with unknown reason"
}

func (r *PipelineRunReconciler) getFailedTasksFromChildReferences(ctx context.Context, pr *v1.PipelineRun) []string {
	var failedTasks []string

	for _, childRef := range pr.Status.ChildReferences {
		// Only look at TaskRuns
		if childRef.Kind == "TaskRun" && childRef.Name != "" {
			// Try to get the TaskRun, extract status for investigation
			if taskRunStatus := r.getTaskRunStatus(ctx, childRef.Name, pr.Namespace); taskRunStatus != nil {
				if r.isTaskRunFailed(taskRunStatus) {
					// Extract reason (if found)
					reason := r.getTaskRunFailureReason(taskRunStatus)
					if reason != "" {
						failedTasks = append(failedTasks, fmt.Sprintf("%s: %s", childRef.PipelineTaskName, reason))
					} else {
						failedTasks = append(failedTasks, fmt.Sprintf("%s: could not determine reason for failure.", childRef.PipelineTaskName))
					}
				}
			}
		}
	}

	return failedTasks
}

func (r *PipelineRunReconciler) getTaskRunStatus(ctx context.Context, taskRunName, namespace string) *v1.TaskRunStatus {
	var taskRun v1.TaskRun
	// Get the TaskRun from the cluster by name and namespace
	err := r.Get(ctx, client.ObjectKey{Name: taskRunName, Namespace: namespace}, &taskRun)
	if err != nil {
		r.Logger.WithError(err).WithFields(logrus.Fields{
			"taskrun":   taskRunName,
			"namespace": namespace,
		}).Debug("Failed to fetch TaskRun details")
		return nil
	}

	return &taskRun.Status
}

func (r *PipelineRunReconciler) isTaskRunFailed(status *v1.TaskRunStatus) bool {
	// If not populated, TaskRun is still in initial phase and not processed yet.
	if status.Conditions == nil {
		return false
	}

	for _, condition := range status.Conditions {
		// Completed but failed
		if condition.Type == "Succeeded" && condition.Status == "False" {
			return true
		}
	}
	return false
}

func (r *PipelineRunReconciler) getTaskRunFailureReason(status *v1.TaskRunStatus) string {
	if status.Conditions == nil {
		return ""
	}

	for _, condition := range status.Conditions {
		// Completed but failed
		if condition.Type == "Succeeded" && condition.Status == "False" {
			// Usually the message has more info. If not, fallback to reason

			if condition.Message != "" {
				return condition.Message
			}
			if condition.Reason != "" {
				return condition.Reason
			}
		}
	}

	// Could not determine reason
	return ""
}

// determineSeverity uses a best-guess approach at determining the severity
// of a failed PipelineRun.
func (r *PipelineRunReconciler) determineSeverity(pr *v1.PipelineRun) string {
	// Name checks

	// Check for indicators that this is for production
	if strings.Contains(pr.Name, "prod") ||
		strings.Contains(pr.Name, "production") {
		return "major"
	}

	// Label checks

	// Check if this is a release
	if serviceType, exists := pr.Labels["appstudio.openshift.io/service"]; exists {
		if serviceType == "release" {
			return "critical"
		}
	}

	// Builds or Tests
	if prType, exists := pr.Labels["pipelines.appstudio.openshift.io/type"]; exists {
		if prType == "build" || prType == "test" {
			return "medium"
		}
	}

	// TODO - figure out what a "low" severity would be.

	// Default
	return "major"
}
