package clients

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type KiteClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

// TODO - These payload structs should probably be exported from Kite service package?
type PipelineFailurePayload struct {
	PipelineName  string `json:"pipelineName"`
	Namespace     string `json:"namespace"`
	FailureReason string `json:"failureReason"`
	RunID         string `json:"runId,omitempty"`
	LogsURL       string `json:"logsUrl,omitempty"`
	Severity      string `json:"severity,omitempty"`
}

type PipelineSuccessPayload struct {
	PipelineName string `json:"pipelineName"`
	Namespace    string `json:"namespace"`
}

// NewKiteClient returns a new client that interacts with the KITE api
func NewKiteClient(baseURL string, logger *logrus.Logger) *KiteClient {
	// Check if we should skip TLS verification (for local development ONLY)
	skipTLS := false
	if skipTLSEnv := os.Getenv("KITE_SKIP_TLS_VERIFY"); skipTLSEnv != "" {
		if skip, err := strconv.ParseBool(skipTLSEnv); err == nil {
			skipTLS = skip
		}
	}

	// Create HTTP client with TLS configurations
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLS,
		},
	}

	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	if skipTLS {
		logger.Warn("TLS certificate verification is disabled for KITE client (KITE_SKIP_TLS_VERIFY=true)")
	}

	return &KiteClient{
		baseURL:    baseURL,
		logger:     logger,
		httpClient: httpClient,
	}
}

// ReportPipelineFailure uses KITE's webhook endpoint for pipeline failures
func (k *KiteClient) ReportPipelineFailure(payload PipelineFailurePayload) error {
	url := fmt.Sprintf("%s/api/v1/webhooks/pipeline-failure", k.baseURL)
	return k.sendWebhook(url, payload, "pipeline-failure")
}

// ReportPipelineSuccess uses KITE's webhook endpoint for pipeline success
func (k *KiteClient) ReportPipelineSuccess(payload PipelineSuccessPayload) error {
	url := fmt.Sprintf("%s/api/v1/webhooks/pipeline-success", k.baseURL)
	return k.sendWebhook(url, payload, "pipeline-success")
}

// sendWebhook is a helper function that sends HTTP requests to KITE
func (k *KiteClient) sendWebhook(url string, payload interface{}, operation string) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		k.logger.WithError(err).Error("Failed to create HTTP request")
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	k.logger.WithFields(logrus.Fields{
		"url":       url,
		"operation": operation,
		"payload":   string(jsonData),
	}).Debug("Sending request to KITE")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		k.logger.WithError(err).Error("Failed to send request to KITE")
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		k.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"operation":   operation,
		}).Errorf("KITE API returned status %d", resp.StatusCode)
		return fmt.Errorf("Error, Status code %d returned", resp.StatusCode)
	}

	k.logger.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"operation":   operation,
	}).Info("Successfully sent request to KITE")

	return nil
}
