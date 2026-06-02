package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// newRemoteWriteList builds a types.List of remote_write objects from the given
// models using the resource's declared object schema. It fails the test on any
// conversion diagnostic so the tests below can focus on the deploy payload.
func newRemoteWriteList(t *testing.T, models []RemoteWriteModel) types.List {
	t.Helper()
	if models == nil {
		// Distinguish "explicit empty list" from "null"; callers that want null
		// should construct it directly.
		models = []RemoteWriteModel{}
	}
	lv, diags := types.ListValueFrom(context.Background(), remoteWriteObjectType(), models)
	if diags.HasError() {
		t.Fatalf("failed to build remote_write list: %v", diags)
	}
	return lv
}

// baseModel returns a model with the non-remote_write/retention fields populated
// the way Create/Update would see them after defaults are applied.
func baseModel() MetricsPrometheusResourceModel {
	return MetricsPrometheusResourceModel{
		Version:        types.StringValue("v2.45.0"),
		Port:           types.Int64Value(9090),
		ScrapeInterval: types.Int64Value(15),
		DeploymentMode: types.StringValue("docker"),
		NetworkMode:    types.StringValue("bridge"),
		RetentionTime:  types.StringValue(""),
		RetentionSize:  types.StringValue(""),
	}
}

func TestBuildDeployRequest_BaseFieldsAlwaysPresent(t *testing.T) {
	r := &MetricsPrometheusResource{}
	data := baseModel()
	data.RemoteWrite = newRemoteWriteList(t, nil) // empty list (the schema default)

	req, diags := r.buildDeployRequest(context.Background(), &data)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	// Base fields must match the backend DeployPrometheusRequest JSON keys.
	if req["prometheus_version"] != "v2.45.0" {
		t.Errorf("prometheus_version = %v, want v2.45.0", req["prometheus_version"])
	}
	if req["prometheus_port"] != int64(9090) {
		t.Errorf("prometheus_port = %v (%T), want int64(9090)", req["prometheus_port"], req["prometheus_port"])
	}
	if req["scrape_interval"] != int64(15) {
		t.Errorf("scrape_interval = %v, want int64(15)", req["scrape_interval"])
	}
	if req["deployment_mode"] != "docker" {
		t.Errorf("deployment_mode = %v, want docker", req["deployment_mode"])
	}
	dc, ok := req["docker_config"].(map[string]interface{})
	if !ok {
		t.Fatalf("docker_config missing or wrong type: %T", req["docker_config"])
	}
	if dc["network_mode"] != "bridge" {
		t.Errorf("docker_config.network_mode = %v, want bridge", dc["network_mode"])
	}
}

func TestBuildDeployRequest_EmptyOptionalFieldsOmitted(t *testing.T) {
	r := &MetricsPrometheusResource{}

	cases := map[string]types.List{
		"empty list (schema default)": newRemoteWriteList(t, nil),
		"null list":                   types.ListNull(remoteWriteObjectType()),
		"unknown list":                types.ListUnknown(remoteWriteObjectType()),
	}

	for name, rw := range cases {
		t.Run(name, func(t *testing.T) {
			data := baseModel()
			data.RemoteWrite = rw

			req, diags := r.buildDeployRequest(context.Background(), &data)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			// Backward-compatibility: when retention is empty and remote_write is
			// empty/null/unknown, none of the new keys may appear in the payload
			// so the API behaves exactly as before the feature existed.
			if _, ok := req["retention_time"]; ok {
				t.Errorf("retention_time should be omitted when empty, got %v", req["retention_time"])
			}
			if _, ok := req["retention_size"]; ok {
				t.Errorf("retention_size should be omitted when empty, got %v", req["retention_size"])
			}
			if _, ok := req["remote_write"]; ok {
				t.Errorf("remote_write should be omitted when empty/null, got %v", req["remote_write"])
			}
		})
	}
}

func TestBuildDeployRequest_RetentionSet(t *testing.T) {
	r := &MetricsPrometheusResource{}
	data := baseModel()
	data.RetentionTime = types.StringValue("30d")
	data.RetentionSize = types.StringValue("50GB")
	data.RemoteWrite = newRemoteWriteList(t, nil)

	req, diags := r.buildDeployRequest(context.Background(), &data)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if req["retention_time"] != "30d" {
		t.Errorf("retention_time = %v, want 30d", req["retention_time"])
	}
	if req["retention_size"] != "50GB" {
		t.Errorf("retention_size = %v, want 50GB", req["retention_size"])
	}
}

func TestBuildDeployRequest_RemoteWriteFullMapping(t *testing.T) {
	r := &MetricsPrometheusResource{}
	data := baseModel()
	data.RemoteWrite = newRemoteWriteList(t, []RemoteWriteModel{
		{
			URL:         types.StringValue("https://prometheus-prod.grafana.net/api/prom/push"),
			Name:        types.StringValue("grafana-cloud"),
			BearerToken: types.StringValue("super-secret-token"),
			BasicAuth: &RemoteWriteBasicAuth{
				Username: types.StringValue("user1"),
				Password: types.StringValue("pass1"),
			},
			TLSConfig: &RemoteWriteTLSConfig{
				CAFile:             types.StringValue("/etc/ssl/ca.pem"),
				CertFile:           types.StringValue("/etc/ssl/cert.pem"),
				KeyFile:            types.StringValue("/etc/ssl/key.pem"),
				ServerName:         types.StringValue("prometheus-prod.grafana.net"),
				InsecureSkipVerify: types.BoolValue(true),
			},
		},
	})

	req, diags := r.buildDeployRequest(context.Background(), &data)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	rwsRaw, ok := req["remote_write"].([]map[string]interface{})
	if !ok {
		t.Fatalf("remote_write missing or wrong type: %T", req["remote_write"])
	}
	if len(rwsRaw) != 1 {
		t.Fatalf("remote_write length = %d, want 1", len(rwsRaw))
	}
	entry := rwsRaw[0]

	if entry["url"] != "https://prometheus-prod.grafana.net/api/prom/push" {
		t.Errorf("url = %v", entry["url"])
	}
	if entry["name"] != "grafana-cloud" {
		t.Errorf("name = %v, want grafana-cloud", entry["name"])
	}
	if entry["bearer_token"] != "super-secret-token" {
		t.Errorf("bearer_token = %v", entry["bearer_token"])
	}

	ba, ok := entry["basic_auth"].(map[string]interface{})
	if !ok {
		t.Fatalf("basic_auth missing or wrong type: %T", entry["basic_auth"])
	}
	if ba["username"] != "user1" || ba["password"] != "pass1" {
		t.Errorf("basic_auth = %v, want user1/pass1", ba)
	}

	// CRITICAL: the tfsdk block is named "tls" but must serialize to the API's
	// "tls_config" key to match the backend RemoteWriteConfig struct.
	if _, ok := entry["tls"]; ok {
		t.Errorf("entry should NOT contain a 'tls' key; the API expects 'tls_config'")
	}
	tlsCfg, ok := entry["tls_config"].(map[string]interface{})
	if !ok {
		t.Fatalf("tls_config missing or wrong type: %T", entry["tls_config"])
	}
	if tlsCfg["ca_file"] != "/etc/ssl/ca.pem" {
		t.Errorf("tls_config.ca_file = %v", tlsCfg["ca_file"])
	}
	if tlsCfg["cert_file"] != "/etc/ssl/cert.pem" {
		t.Errorf("tls_config.cert_file = %v", tlsCfg["cert_file"])
	}
	if tlsCfg["key_file"] != "/etc/ssl/key.pem" {
		t.Errorf("tls_config.key_file = %v", tlsCfg["key_file"])
	}
	if tlsCfg["server_name"] != "prometheus-prod.grafana.net" {
		t.Errorf("tls_config.server_name = %v", tlsCfg["server_name"])
	}
	if tlsCfg["insecure_skip_verify"] != true {
		t.Errorf("tls_config.insecure_skip_verify = %v, want true", tlsCfg["insecure_skip_verify"])
	}
}

func TestBuildDeployRequest_RemoteWriteMinimalOmitsEmptyAuth(t *testing.T) {
	r := &MetricsPrometheusResource{}
	data := baseModel()
	// A single endpoint with only a URL; basic_auth/tls are nil and the optional
	// scalar fields are empty. None of the empty sub-objects/fields should appear.
	data.RemoteWrite = newRemoteWriteList(t, []RemoteWriteModel{
		{
			URL:         types.StringValue("https://remote.example/push"),
			Name:        types.StringValue(""),
			BearerToken: types.StringValue(""),
		},
	})

	req, diags := r.buildDeployRequest(context.Background(), &data)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	rwsRaw := req["remote_write"].([]map[string]interface{})
	entry := rwsRaw[0]
	if entry["url"] != "https://remote.example/push" {
		t.Errorf("url = %v", entry["url"])
	}
	for _, k := range []string{"name", "bearer_token", "basic_auth", "tls_config"} {
		if _, ok := entry[k]; ok {
			t.Errorf("empty optional %q should be omitted, got %v", k, entry[k])
		}
	}
}

func TestBuildDeployRequest_RemoteWriteInsecureSkipVerifyFalseOmitted(t *testing.T) {
	r := &MetricsPrometheusResource{}
	data := baseModel()
	data.RemoteWrite = newRemoteWriteList(t, []RemoteWriteModel{
		{
			URL: types.StringValue("https://remote.example/push"),
			TLSConfig: &RemoteWriteTLSConfig{
				ServerName:         types.StringValue("remote.example"),
				InsecureSkipVerify: types.BoolValue(false),
			},
		},
	})

	req, _ := r.buildDeployRequest(context.Background(), &data)
	entry := req["remote_write"].([]map[string]interface{})[0]
	tlsCfg, ok := entry["tls_config"].(map[string]interface{})
	if !ok {
		t.Fatalf("tls_config missing: %T", entry["tls_config"])
	}
	// insecure_skip_verify=false matches the API omitempty default, so it must
	// not be serialized (avoids a spurious "true vs absent" difference).
	if _, ok := tlsCfg["insecure_skip_verify"]; ok {
		t.Errorf("insecure_skip_verify=false should be omitted, got %v", tlsCfg["insecure_skip_verify"])
	}
	if tlsCfg["server_name"] != "remote.example" {
		t.Errorf("server_name = %v", tlsCfg["server_name"])
	}
}

// remoteWriteObjectType must stay in sync with the model struct tags; ElementsAs
// in buildDeployRequest depends on that alignment. This is implicitly exercised
// above, but assert the attribute set explicitly to catch accidental drift.
func TestRemoteWriteObjectTypeAttributes(t *testing.T) {
	ot := remoteWriteObjectType()
	want := map[string]attr.Type{
		"url":          types.StringType,
		"name":         types.StringType,
		"bearer_token": types.StringType,
	}
	for k, typ := range want {
		got, ok := ot.AttrTypes[k]
		if !ok {
			t.Errorf("remoteWriteObjectType missing attr %q", k)
			continue
		}
		if got != typ {
			t.Errorf("attr %q type = %v, want %v", k, got, typ)
		}
	}
	if _, ok := ot.AttrTypes["basic_auth"]; !ok {
		t.Errorf("remoteWriteObjectType missing basic_auth")
	}
	if _, ok := ot.AttrTypes["tls"]; !ok {
		t.Errorf("remoteWriteObjectType missing tls")
	}
}

// TestReadStatus_PreservesWriteOnlyFields asserts that Read (via readStatus)
// does NOT clobber retention_time/retention_size/remote_write, which the
// /metrics/status endpoint never echoes back. Overwriting them would cause a
// perpetual diff and "inconsistent result" errors.
func TestReadStatus_PreservesWriteOnlyFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/metrics/status" {
			w.Header().Set("Content-Type", "application/json")
			// Deliberately does NOT include retention/remote_write fields.
			_, _ = w.Write([]byte(`{
				"version": "v2.45.0",
				"port": 9090,
				"status": "running",
				"started_at": "2025-01-01T00:00:00Z",
				"deployment_mode": "docker",
				"network_mode": "bridge"
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	r := &MetricsPrometheusResource{client: NewClient(server.URL, "", "user", "pass")}

	// Simulate prior state carrying the write-only fields.
	data := baseModel()
	data.RetentionTime = types.StringValue("90d")
	data.RetentionSize = types.StringValue("100GB")
	data.RemoteWrite = newRemoteWriteList(t, []RemoteWriteModel{
		{URL: types.StringValue("https://remote.example/push")},
	})

	if err := r.readStatus(context.Background(), &data); err != nil {
		t.Fatalf("readStatus error: %v", err)
	}

	// Status-derived fields should be refreshed.
	if data.Status.ValueString() != "running" {
		t.Errorf("status = %q, want running", data.Status.ValueString())
	}
	if data.StartedAt.ValueString() != "2025-01-01T00:00:00Z" {
		t.Errorf("started_at = %q", data.StartedAt.ValueString())
	}

	// Write-only fields must be preserved untouched.
	if data.RetentionTime.ValueString() != "90d" {
		t.Errorf("retention_time = %q, want 90d (preserved)", data.RetentionTime.ValueString())
	}
	if data.RetentionSize.ValueString() != "100GB" {
		t.Errorf("retention_size = %q, want 100GB (preserved)", data.RetentionSize.ValueString())
	}
	if data.RemoteWrite.IsNull() || len(data.RemoteWrite.Elements()) != 1 {
		t.Errorf("remote_write should be preserved with 1 element, got null=%v len=%d",
			data.RemoteWrite.IsNull(), len(data.RemoteWrite.Elements()))
	}
}
