# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - TBD

### Added

#### Core Resources
- **Organizations**: Manage Hyperledger Fabric organizations with MSP configuration
- **Key Providers**: Support for Database, AWS KMS, and HashiCorp Vault key storage
- **Cryptographic Keys**: Create and manage RSA, EC (NIST + secp256k1), and ED25519 keys

#### Fabric Resources
- **Fabric Peers**: Deploy and configure Hyperledger Fabric peer nodes
- **Fabric Orderers**: Deploy and configure Fabric orderer nodes with Raft consensus
- **Fabric Networks**: Create Fabric channels and configure network topology
- **Fabric Join Nodes**: Join peers and orderers to channels
- **Fabric Anchor Peers**: Configure anchor peers for organizations
- **Fabric Identities**: Generate admin/client identities with X.509 certificates

#### Fabric Chaincode Lifecycle
- **Chaincode**: Create chaincode records associated with networks
- **Chaincode Definitions**: Define versions, sequences, docker images, and endorsement policies
- **Chaincode Install**: Install chaincode docker images on peer nodes
- **Chaincode Approve**: Organization-level chaincode approval
- **Chaincode Commit**: Commit chaincode definitions to channels
- **Chaincode Deploy**: Deploy and start chaincode containers with environment variables

#### Besu Resources
- **Besu Networks**: Create Hyperledger Besu networks with QBFT/IBFT2 consensus
- **Besu Nodes**: Deploy Besu validator and transaction nodes

#### Backup & Recovery
- **Backup Targets**: Configure S3-compatible storage for backups (AWS S3, MinIO, etc.)
- **Backup Schedules**: Automated cron-based backup schedules with retention policies

#### Notifications
- **Notification Providers**: SMTP email notifications for backup events, node downtime, and S3 issues

#### Plugins
- **Plugin Definitions**: Register plugins from YAML specifications with Docker Compose
- **Plugin Deployments**: Deploy plugins with parameterized configurations

#### Data Sources
- Data sources for all resources to query existing infrastructure
- Support for filtering by ID, name, MSP ID, and other attributes

### Features
- Complete Hyperledger Fabric network lifecycle management
- Docker image-based chaincode deployment (no package files required)
- Automated Vault initialization and status monitoring
- AWS KMS integration with LocalStack support
- secp256k1 curve support for Ethereum/Besu keys
- Idempotent chaincode operations (install/approve/commit)
- Comprehensive error handling and validation
- Type-safe nested configuration objects
- Automatic certificate generation for Fabric identities
- Plugin system for platform extensibility
- Prometheus metrics integration for plugins
- S3-compatible backup with Restic encryption

### Documentation
- Complete resource documentation with examples
- Step-by-step setup guides for Fabric and Besu networks
- Key provider configuration examples for all backends
- Chaincode lifecycle workflow documentation
- Plugin development and deployment guides
- Backup and restore procedures
- Troubleshooting guides for common issues

### Examples
- Database key provider examples for all algorithms
- AWS KMS with LocalStack integration
- HashiCorp Vault CREATE and IMPORT modes
- Complete Fabric network with peers, orderers, and channels
- Full chaincode lifecycle (install → approve → commit → deploy)
- Besu network with QBFT consensus and multiple validators
- MinIO backup configuration with automated schedules
- Email notifications with Mailpit SMTP server
- HLF API plugin registration and deployment

[Unreleased]: https://github.com/kfsoftware/terraform-provider-chainlaunch/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/kfsoftware/terraform-provider-chainlaunch/releases/tag/v0.1.0
