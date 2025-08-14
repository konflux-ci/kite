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
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	knative "knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	KiteBridgeOperatorNamespace = "kite-bridge-operator"
)

func setupPipelineRun(name string, options PipelineRunBuilderOptions) {
	builder := NewPipelineRunBuilder(name, KiteBridgeOperatorNamespace)
	var pipelineRun *v1.PipelineRun
	if options.Labels != nil {
		builder.WithLabels(options.Labels)
	}

	pipelineRun = builder.Build()
	Expect(k8sClient.Create(ctx, pipelineRun)).Should(Succeed())
	current := &v1.PipelineRun{}
	key := types.NamespacedName{Name: name, Namespace: KiteBridgeOperatorNamespace}

	Eventually(func(g Gomega) {
		g.Expect(k8sClient.Get(ctx, key, current)).To(Succeed())
	}).Should(Succeed())

	// Things like Status need to be updated after
	if options.Conditions != nil {
		current.Status.Conditions = options.Conditions
	}
	if options.CompletionTime != nil {
		current.Status.CompletionTime = options.CompletionTime
	}

	Eventually(func(g Gomega) {
		g.Expect(k8sClient.Status().Update(ctx, current)).To(Succeed())
	}).Should(Succeed())
}

func tearDownPipelineRuns() {
	pipelineRuns := listPipelineRuns(KiteBridgeOperatorNamespace)
	for _, pipelineRuns := range pipelineRuns {
		Expect(k8sClient.Delete(ctx, &pipelineRuns)).Should(Succeed())
	}
	Eventually(func() []v1.PipelineRun {
		return listPipelineRuns(KiteBridgeOperatorNamespace)
	}).Should(HaveLen(0))
}

var _ = Describe("PipelineRun Controller", func() {
	var (
		reconciler     *PipelineRunReconciler
		mockKiteClient *MockKiteClient
		logBuffer      bytes.Buffer
		logger         *logrus.Logger
	)

	BeforeEach(func() {
		createNamespace(KiteBridgeOperatorNamespace)
		mockKiteClient = &MockKiteClient{}
		logger = logrus.New()
		logger.SetOutput(&logBuffer)

		reconciler = &PipelineRunReconciler{
			Client:     k8sClient,
			Scheme:     k8sClient.Scheme(),
			KiteClient: mockKiteClient,
			Logger:     logger,
		}
	})

	AfterEach(func() {
		logBuffer.Reset()
		tearDownPipelineRuns()
	})

	Context("When a PipelineRun fails", func() {
		var (
			prName    = "failed-pipeline-run"
			lookupKey = types.NamespacedName{Name: prName, Namespace: KiteBridgeOperatorNamespace}
			pr        = &v1.PipelineRun{}
		)

		BeforeEach(func() {
			now := metav1.Now()
			setupPipelineRun(prName, PipelineRunBuilderOptions{
				Conditions: []knative.Condition{
					{
						Type:    "Succeeded",
						Message: "Tasks Completed: 1 (Failed: 0, Cancelled: 0), Skipped: 0",
						Status:  "False",
						Reason:  "Failed",
					},
				},
				Labels: map[string]string{
					"tekton.dev/pipeline": "simple-pipeline",
				},
				CompletionTime: &now,
			})

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, lookupKey, pr)).To(Succeed())
			}).Should(Succeed())
		})

		It("should successfully reconcile the PipelineRun when it fails", func() {
			// Trigger reconciliation
			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: lookupKey,
			})

			// Verify reconciliation succeeded
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			// Verify Kite client was called with failure
			Expect(mockKiteClient.FailureReports).To(HaveLen(1))
			failureReport := mockKiteClient.FailureReports[0]

			// Verify PR has what we expect
			Expect(failureReport.PipelineName).To(Equal("simple-pipeline"))
			Expect(failureReport.Namespace).To(Equal(KiteBridgeOperatorNamespace))
			Expect(failureReport.FailureReason).To(ContainSubstring("Tasks Completed"))
			Expect(failureReport.RunID).To(Equal(string(pr.UID)))
			Expect(failureReport.Severity).To(Equal("major"))
		})
	})
})
