package controller

import (
	"context"
	"fmt"

	"github.com/konflux-ci/kite/packages/operator/internal/clients"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knative "knative.dev/pkg/apis/duck/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PipelineRunBuilder struct {
	name      string
	namespace string
	pr        *v1.PipelineRun
}

type PipelineRunBuilderOptions struct {
	Labels         map[string]string
	Conditions     knative.Conditions
	CompletionTime *metav1.Time
}

func NewPipelineRunBuilder(name, namespace string) *PipelineRunBuilder {
	return &PipelineRunBuilder{
		name:      name,
		namespace: namespace,
		pr: &v1.PipelineRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
	}
}

func (b *PipelineRunBuilder) Build() *v1.PipelineRun {
	return b.pr
}

func (b *PipelineRunBuilder) WithAnnotations(annotations map[string]string) *PipelineRunBuilder {
	if b.pr.Annotations == nil {
		b.pr.Annotations = make(map[string]string)
	}
	for key, value := range annotations {
		b.pr.Annotations[key] = value
	}
	return b
}

func (b *PipelineRunBuilder) WithLabels(labels map[string]string) *PipelineRunBuilder {
	if b.pr.Labels == nil {
		b.pr.Labels = make(map[string]string)
	}
	for k, v := range labels {
		b.pr.Labels[k] = v
	}
	return b
}

func (b *PipelineRunBuilder) WithConditions(conditions knative.Conditions) *PipelineRunBuilder {
	b.pr.Status.Conditions = conditions
	return b
}

func (b *PipelineRunBuilder) WithCompletionTime(time metav1.Time) *PipelineRunBuilder {
	b.pr.Status.CompletionTime = &time
	return b
}

func listPipelineRuns(namespace string) []v1.PipelineRun {
	pipelineRuns := &v1.PipelineRunList{}
	_ = k8sClient.List(ctx, pipelineRuns, client.InNamespace(namespace))
	return pipelineRuns.Items
}

func createNamespace(name string) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	_ = k8sClient.Create(ctx, ns)
}

type MockKiteClient struct {
	FailureReports []clients.PipelineFailurePayload
	SuccessReports []clients.PipelineSuccessPayload
	ShouldFail     bool
}

// Ensure we're implementing the interface
var _ clients.KiteWebhookClient = (*MockKiteClient)(nil)

func (m *MockKiteClient) ReportPipelineFailure(ctx context.Context, payload clients.PipelineFailurePayload) error {
	m.FailureReports = append(m.FailureReports, payload)
	if m.ShouldFail {
		return fmt.Errorf("Failed to report pipeline failure")
	}
	return nil
}

func (m *MockKiteClient) ReportPipelineSuccess(ctx context.Context, payload clients.PipelineSuccessPayload) error {
	m.SuccessReports = append(m.SuccessReports, payload)
	if m.ShouldFail {
		return fmt.Errorf("failed to report pipeline success")
	}
	return nil
}
