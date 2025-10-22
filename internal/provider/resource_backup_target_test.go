package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccBackupTargetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccBackupTargetResourceConfig("MinIO Test", "http://localhost:9000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("chainlaunch_backup_target.test", "name", "MinIO Test"),
					resource.TestCheckResourceAttr("chainlaunch_backup_target.test", "type", "S3"),
					resource.TestCheckResourceAttr("chainlaunch_backup_target.test", "endpoint", "http://localhost:9000"),
					resource.TestCheckResourceAttr("chainlaunch_backup_target.test", "region", "us-east-1"),
					resource.TestCheckResourceAttr("chainlaunch_backup_target.test", "bucket_name", "test-backups"),
					resource.TestCheckResourceAttr("chainlaunch_backup_target.test", "force_path_style", "true"),
					resource.TestCheckResourceAttrSet("chainlaunch_backup_target.test", "id"),
					resource.TestCheckResourceAttrSet("chainlaunch_backup_target.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "chainlaunch_backup_target.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_access_key", "restic_password"},
			},
			// Update and Read testing
			{
				Config: testAccBackupTargetResourceConfig("MinIO Updated", "http://localhost:9000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("chainlaunch_backup_target.test", "name", "MinIO Updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccBackupTargetResourceConfig(name, endpoint string) string {
	return fmt.Sprintf(`
resource "chainlaunch_backup_target" "test" {
  name               = %[1]q
  type               = "S3"
  endpoint           = %[2]q
  region             = "us-east-1"
  access_key_id      = "test-access-key"
  secret_access_key  = "test-secret-key"
  bucket_name        = "test-backups"
  bucket_path        = "terraform-test"
  force_path_style   = true
  restic_password    = "test-restic-password"
}
`, name, endpoint)
}

// Unit tests with mock server
func TestBackupTargetResourceWithMockServer(t *testing.T) {
	// Create a mock storage for backup targets
	storage := &mockBackupTargetStorage{
		targets: make(map[int]BackupTarget),
		nextID:  1,
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/backups/targets":
			storage.handleCreate(w, r)
		case r.Method == "GET" && r.URL.Path == "/backups/targets/1":
			storage.handleGet(w, r, 1)
		case r.Method == "PUT" && r.URL.Path == "/backups/targets/1":
			storage.handleUpdate(w, r, 1)
		case r.Method == "DELETE" && r.URL.Path == "/backups/targets/1":
			storage.handleDelete(w, r, 1)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupTargetMockConfig(server.URL, "Test Target"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBackupTargetExists("chainlaunch_backup_target.mock_test", storage),
					resource.TestCheckResourceAttr("chainlaunch_backup_target.mock_test", "name", "Test Target"),
					resource.TestCheckResourceAttr("chainlaunch_backup_target.mock_test", "type", "S3"),
				),
			},
		},
	})
}

func testAccBackupTargetMockConfig(providerURL, name string) string {
	return fmt.Sprintf(`
provider "chainlaunch" {
  url      = %[1]q
  username = "test"
  password = "test"
}

resource "chainlaunch_backup_target" "mock_test" {
  name               = %[2]q
  type               = "S3"
  endpoint           = "http://localhost:9000"
  region             = "us-east-1"
  access_key_id      = "mock-key"
  secret_access_key  = "mock-secret"
  bucket_name        = "mock-bucket"
  force_path_style   = true
  restic_password    = "mock-password"
}
`, providerURL, name)
}

func testAccCheckBackupTargetExists(resourceName string, storage *mockBackupTargetStorage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Invalid ID: %s", rs.Primary.ID)
		}

		storage.mu.Lock()
		defer storage.mu.Unlock()

		if _, ok := storage.targets[id]; !ok {
			return fmt.Errorf("Backup target not found in storage")
		}

		return nil
	}
}

// Mock storage for backup targets
type mockBackupTargetStorage struct {
	mu      sync.Mutex
	targets map[int]BackupTarget
	nextID  int
}

func (s *mockBackupTargetStorage) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateBackupTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	target := BackupTarget{
		ID:             s.nextID,
		Name:           req.Name,
		Type:           req.Type,
		Endpoint:       req.Endpoint,
		Region:         req.Region,
		AccessKeyID:    req.AccessKeyID,
		BucketName:     req.BucketName,
		BucketPath:     req.BucketPath,
		ForcePathStyle: req.ForcePathStyle,
		CreatedAt:      "2025-01-01T00:00:00Z",
		UpdatedAt:      "2025-01-01T00:00:00Z",
	}

	s.targets[s.nextID] = target
	s.nextID++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(target)
}

func (s *mockBackupTargetStorage) handleGet(w http.ResponseWriter, r *http.Request, id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	target, ok := s.targets[id]
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(target)
}

func (s *mockBackupTargetStorage) handleUpdate(w http.ResponseWriter, r *http.Request, id int) {
	var req CreateBackupTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	target, ok := s.targets[id]
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	target.Name = req.Name
	target.Type = req.Type
	target.Endpoint = req.Endpoint
	target.Region = req.Region
	target.AccessKeyID = req.AccessKeyID
	target.BucketName = req.BucketName
	target.BucketPath = req.BucketPath
	target.ForcePathStyle = req.ForcePathStyle
	target.UpdatedAt = "2025-01-01T00:00:01Z"

	s.targets[id] = target

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(target)
}

func (s *mockBackupTargetStorage) handleDelete(w http.ResponseWriter, r *http.Request, id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.targets[id]; !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	delete(s.targets, id)
	w.WriteHeader(http.StatusNoContent)
}
