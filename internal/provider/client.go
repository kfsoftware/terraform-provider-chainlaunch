package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client is the Chainlaunch API client
type Client struct {
	BaseURL    string
	APIKey     string
	Username   string
	Password   string
	HTTPClient *http.Client
}

// NewClient creates a new Chainlaunch API client
// Supports both API key and username/password authentication
func NewClient(baseURL, apiKey, username, password string) *Client {
	return &Client{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Username: username,
		Password: password,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

// DoRequest performs an HTTP request to the Chainlaunch API
func (c *Client) DoRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	var jsonBody []byte

	if body != nil {
		var err error
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s/api/v1%s", c.BaseURL, path)
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set authentication - prefer username/password if provided, otherwise use API key
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	} else if c.APIKey != "" {
		req.SetBasicAuth(c.APIKey, "")
	}

	// Debug logging if TF_LOG is set
	if os.Getenv("TF_LOG") != "" {
		log.Printf("[DEBUG] Chainlaunch API Request: %s %s", method, url)
		if jsonBody != nil {
			log.Printf("[DEBUG] Request Body: %s", string(jsonBody))
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Explicitly ignore error on close
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Debug logging if TF_LOG is set
	if os.Getenv("TF_LOG") != "" {
		log.Printf("[DEBUG] Chainlaunch API Response: Status %d", resp.StatusCode)
		log.Printf("[DEBUG] Response Body: %s", string(respBody))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// For 404, return a special error that can be detected by resources
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("NOT_FOUND: %s", string(respBody))
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// IsNotFoundError checks if an error is a 404 Not Found error
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "NOT_FOUND:")
}

// Organization types
type Organization struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	MSPID       string `json:"mspId"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	MSPID       string `json:"mspId"`
	Description string `json:"description,omitempty"`
	ProviderID  int    `json:"providerId,omitempty"`
}

// Node types
type Node struct {
	ID        int64                  `json:"id"`
	Name      string                 `json:"name"`
	Platform  string                 `json:"platform"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status,omitempty"`
	CreatedAt string                 `json:"createdAt,omitempty"`
	UpdatedAt string                 `json:"updatedAt,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

type CreateNodeRequest struct {
	Name     string                 `json:"name"`
	Platform string                 `json:"platform"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// Network types
type Network struct {
	ID        int64                  `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status,omitempty"`
	CreatedAt string                 `json:"createdAt,omitempty"`
	UpdatedAt string                 `json:"updatedAt,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

type CreateNetworkRequest struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// KeyProvider types
type KeyProvider struct {
	ID        int64                  `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status,omitempty"`
	CreatedAt string                 `json:"createdAt,omitempty"`
	UpdatedAt string                 `json:"updatedAt,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

type CreateKeyProviderRequest struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// Fabric Network types
type FabricNetworkResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Platform    string `json:"platform"`
	Status      string `json:"status,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type CreateFabricNetworkRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Config      FabricNetworkConfig `json:"config"`
}

type UpdateFabricNetworkRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Config      FabricNetworkConfig `json:"config"`
}

type FabricNetworkConfig struct {
	PeerOrganizations       []OrganizationConfig    `json:"peerOrganizations,omitempty"`
	OrdererOrganizations    []OrganizationConfig    `json:"ordererOrganizations,omitempty"`
	ExternalPeerOrgs        []ExternalOrgConfig     `json:"externalPeerOrgs,omitempty"`
	ExternalOrdererOrgs     []ExternalOrgConfig     `json:"externalOrdererOrgs,omitempty"`
	ConsensusType           string                  `json:"consensusType,omitempty"`
	EtcdRaftOptions         *EtcdRaftOptions        `json:"etcdRaftOptions,omitempty"`
	SmartBFTOptions         *SmartBFTOptions        `json:"smartBFTOptions,omitempty"`
	SmartBFTConsenters      []SmartBFTConsenter     `json:"smartBFTConsenters,omitempty"`
	ChannelCapabilities     []string                `json:"channelCapabilities,omitempty"`
	ApplicationCapabilities []string                `json:"applicationCapabilities,omitempty"`
	OrdererCapabilities     []string                `json:"ordererCapabilities,omitempty"`
	BatchSize               *BatchSize              `json:"batchSize,omitempty"`
	BatchTimeout            string                  `json:"batchTimeout,omitempty"`
	ApplicationPolicies     map[string]FabricPolicy `json:"applicationPolicies,omitempty"`
	OrdererPolicies         map[string]FabricPolicy `json:"ordererPolicies,omitempty"`
	ChannelPolicies         map[string]FabricPolicy `json:"channelPolicies,omitempty"`
}

type OrganizationConfig struct {
	ID      int64   `json:"id"`
	NodeIDs []int64 `json:"nodeIds"`
}

type ExternalOrgConfig struct {
	MSPID      string            `json:"mspid"`
	SignCACert string            `json:"signCACert"`
	TLSCACert  string            `json:"tlsCACert"`
	Consenters []ConsenterConfig `json:"consenters,omitempty"`
}

type ConsenterConfig struct {
	Host    string `json:"host"`
	Port    int64  `json:"port"`
	TLSCert string `json:"tlsCert,omitempty"`
}

type BatchSize struct {
	MaxMessageCount   int `json:"maxMessageCount,omitempty"`
	AbsoluteMaxBytes  int `json:"absoluteMaxBytes,omitempty"`
	PreferredMaxBytes int `json:"preferredMaxBytes,omitempty"`
}

type EtcdRaftOptions struct {
	TickInterval         string `json:"tickInterval,omitempty"`
	ElectionTick         int    `json:"electionTick,omitempty"`
	HeartbeatTick        int    `json:"heartbeatTick,omitempty"`
	MaxInflightBlocks    int    `json:"maxInflightBlocks,omitempty"`
	SnapshotIntervalSize int    `json:"snapshotIntervalSize,omitempty"`
}

type SmartBFTOptions struct {
	RequestBatchMaxCount      int    `json:"requestBatchMaxCount,omitempty"`
	RequestBatchMaxBytes      int    `json:"requestBatchMaxBytes,omitempty"`
	RequestBatchMaxInterval   string `json:"requestBatchMaxInterval,omitempty"`
	RequestMaxBytes           int    `json:"requestMaxBytes,omitempty"`
	IncomingMessageBufferSize int    `json:"incomingMessageBufferSize,omitempty"`
	RequestPoolSize           int    `json:"requestPoolSize,omitempty"`
	ViewChangeResendInterval  string `json:"viewChangeResendInterval,omitempty"`
	ViewChangeTimeout         string `json:"viewChangeTimeout,omitempty"`
	LeaderHeartbeatCount      int    `json:"leaderHeartbeatCount,omitempty"`
	LeaderHeartbeatTimeout    string `json:"leaderHeartbeatTimeout,omitempty"`
	CollectTimeout            string `json:"collectTimeout,omitempty"`
	SyncOnStart               bool   `json:"syncOnStart,omitempty"`
	SpeedUpViewChange         bool   `json:"speedUpViewChange,omitempty"`
	LeaderRotation            string `json:"leaderRotation,omitempty"`
	DecisionsPerLeader        int    `json:"decisionsPerLeader,omitempty"`
	RequestComplainTimeout    string `json:"requestComplainTimeout,omitempty"`
	RequestAutoRemoveTimeout  string `json:"requestAutoRemoveTimeout,omitempty"`
	RequestForwardTimeout     string `json:"requestForwardTimeout,omitempty"`
}

type SmartBFTConsenter struct {
	ID            int64    `json:"id"`
	MSPID         string   `json:"mspId"`
	Identity      string   `json:"identity"`
	ClientTLSCert string   `json:"clientTLSCert"`
	ServerTLSCert string   `json:"serverTLSCert"`
	Address       HostPort `json:"address"`
}

type HostPort struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type FabricPolicy struct {
	Type string `json:"type"`
	Rule string `json:"rule"`
}

// Join node to network types
type AddNodeToNetworkRequest struct {
	NodeID int64  `json:"nodeId"`
	Role   string `json:"role"`
}

type GetNetworkNodesResponse struct {
	Nodes []NetworkNode `json:"nodes"`
}

type NetworkNode struct {
	NodeID int64  `json:"nodeId"`
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
}

// Backup Target types
type BackupTarget struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Endpoint       string `json:"endpoint,omitempty"`
	Region         string `json:"region"`
	AccessKeyID    string `json:"accessKeyId"`
	BucketName     string `json:"bucketName"`
	BucketPath     string `json:"bucketPath,omitempty"`
	ForcePathStyle bool   `json:"forcePathStyle"`
	ResticPassword string `json:"resticPassword,omitempty"`
	CreatedAt      string `json:"createdAt,omitempty"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
}

type CreateBackupTargetRequest struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	Endpoint        string `json:"endpoint,omitempty"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretKey"` // API expects "secretKey" not "secretAccessKey"
	BucketName      string `json:"bucketName"`
	BucketPath      string `json:"bucketPath,omitempty"`
	ForcePathStyle  bool   `json:"forcePathStyle"`
	ResticPassword  string `json:"resticPassword"`
}

// Backup Schedule types
type BackupSchedule struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	TargetID       int    `json:"targetId"`
	CronExpression string `json:"cronExpression"`
	Enabled        bool   `json:"enabled"`
	RetentionDays  int    `json:"retentionDays"`
	LastRunAt      string `json:"lastRunAt,omitempty"`
	NextRunAt      string `json:"nextRunAt,omitempty"`
	CreatedAt      string `json:"createdAt,omitempty"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
}

type CreateBackupScheduleRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	TargetID       int    `json:"targetId"`
	CronExpression string `json:"cronExpression"`
	Enabled        bool   `json:"enabled"`
	RetentionDays  int    `json:"retentionDays"`
}

// Backup types
type Backup struct {
	ID           int         `json:"id"`
	TargetID     int         `json:"targetId"`
	ScheduleID   int         `json:"scheduleId,omitempty"`
	Status       string      `json:"status"`
	SizeBytes    int64       `json:"sizeBytes,omitempty"`
	Metadata     interface{} `json:"metadata,omitempty"`
	StartedAt    string      `json:"startedAt,omitempty"`
	CompletedAt  string      `json:"completedAt,omitempty"`
	ErrorMessage string      `json:"errorMessage,omitempty"`
	CreatedAt    string      `json:"createdAt,omitempty"`
}

type CreateBackupRequest struct {
	TargetID   int         `json:"targetId"`
	ScheduleID int         `json:"scheduleId,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

// Node Invitation types
type NodeInvitationRequest struct {
	Bidirectional bool                   `json:"bidirectional,omitempty"`
	Options       map[string]interface{} `json:"options,omitempty"`
}

type NodeInvitationResponse struct {
	InvitationJWT string `json:"invitation_jwt"`
}

type NodeAcceptInvitationRequest struct {
	InvitationJWT string `json:"invitation_jwt"`
}

type NodeAcceptInvitationResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
