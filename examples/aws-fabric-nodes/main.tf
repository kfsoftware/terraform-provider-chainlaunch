terraform {
  required_providers {
    chainlaunch = {
      source = "registry.terraform.io/kfsoftware/chainlaunch"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# ==============================================================================
# AWS PROVIDER CONFIGURATION
# ==============================================================================

provider "aws" {
  region = var.aws_region
}

# Install Chainlaunch on a management server (optional - if you don't have an existing instance)
resource "chainlaunch_install_ssh" "management_server" {
  count = var.install_chainlaunch_server ? 1 : 0

  host        = aws_instance.chainlaunch_management[0].public_ip
  user        = "ubuntu"
  private_key = file(var.ssh_private_key_path)

  version      = var.chainlaunch_version
  install_path = "/opt/chainlaunch"
  data_path    = "/var/lib/chainlaunch"
  port_8100    = 8100

  auto_start = true

  environment = {
    LOG_LEVEL   = "info"
    TZ          = "UTC"
    DOCKER_HOST = "unix:///var/run/docker.sock"
  }

  depends_on = [aws_instance.chainlaunch_management]
}

# Configure the Chainlaunch Provider
# If install_chainlaunch_server is true, use the installed server, otherwise use the provided URL
provider "chainlaunch" {
  url      = var.install_chainlaunch_server ? chainlaunch_install_ssh.management_server[0].chainlaunch_url : var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# ==============================================================================
# DATA SOURCES
# ==============================================================================

# Get the latest Ubuntu AMI
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Fetch the default key provider for Chainlaunch
data "chainlaunch_key_providers" "default" {}

# ==============================================================================
# AWS NETWORKING
# ==============================================================================

# Create VPC
resource "aws_vpc" "fabric_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.project_name}-vpc"
    Environment = var.environment
  }
}

# Create Internet Gateway
resource "aws_internet_gateway" "fabric_igw" {
  vpc_id = aws_vpc.fabric_vpc.id

  tags = {
    Name        = "${var.project_name}-igw"
    Environment = var.environment
  }
}

# Create Public Subnet
resource "aws_subnet" "fabric_public_subnet" {
  vpc_id                  = aws_vpc.fabric_vpc.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name        = "${var.project_name}-public-subnet"
    Environment = var.environment
  }
}

# Get available AZs
data "aws_availability_zones" "available" {
  state = "available"
}

# Create Route Table
resource "aws_route_table" "fabric_public_rt" {
  vpc_id = aws_vpc.fabric_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.fabric_igw.id
  }

  tags = {
    Name        = "${var.project_name}-public-rt"
    Environment = var.environment
  }
}

# Associate Route Table with Subnet
resource "aws_route_table_association" "fabric_public_rta" {
  subnet_id      = aws_subnet.fabric_public_subnet.id
  route_table_id = aws_route_table.fabric_public_rt.id
}

# ==============================================================================
# AWS SECURITY GROUPS
# ==============================================================================

# Security Group for Fabric Peers
resource "aws_security_group" "fabric_peer_sg" {
  name        = "${var.project_name}-peer-sg"
  description = "Security group for Hyperledger Fabric peer nodes"
  vpc_id      = aws_vpc.fabric_vpc.id

  # Peer port (7051)
  ingress {
    description = "Fabric Peer"
    from_port   = 7051
    to_port     = 7051
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Chaincode port (7052)
  ingress {
    description = "Fabric Chaincode"
    from_port   = 7052
    to_port     = 7052
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Events port (7053)
  ingress {
    description = "Fabric Events"
    from_port   = 7053
    to_port     = 7053
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Operations/Metrics port (9443)
  ingress {
    description = "Operations/Metrics"
    from_port   = 9443
    to_port     = 9443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # SSH
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow all outbound
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.project_name}-peer-sg"
    Environment = var.environment
  }
}

# Security Group for Fabric Orderers
resource "aws_security_group" "fabric_orderer_sg" {
  name        = "${var.project_name}-orderer-sg"
  description = "Security group for Hyperledger Fabric orderer nodes"
  vpc_id      = aws_vpc.fabric_vpc.id

  # Orderer port (7050)
  ingress {
    description = "Fabric Orderer"
    from_port   = 7050
    to_port     = 7050
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Operations/Metrics port (9443)
  ingress {
    description = "Operations/Metrics"
    from_port   = 9443
    to_port     = 9443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # SSH
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow all outbound
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.project_name}-orderer-sg"
    Environment = var.environment
  }
}

# ==============================================================================
# AWS EC2 INSTANCES - CHAINLAUNCH MANAGEMENT SERVER (OPTIONAL)
# ==============================================================================

# Security Group for Chainlaunch Management Server
resource "aws_security_group" "chainlaunch_management_sg" {
  count = var.install_chainlaunch_server ? 1 : 0

  name        = "${var.project_name}-chainlaunch-sg"
  description = "Security group for Chainlaunch management server"
  vpc_id      = aws_vpc.fabric_vpc.id

  # Chainlaunch API port (8100)
  ingress {
    description = "Chainlaunch API"
    from_port   = 8100
    to_port     = 8100
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # SSH
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow all outbound
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.project_name}-chainlaunch-sg"
    Environment = var.environment
  }
}

# Create EC2 instance for Chainlaunch management server
resource "aws_instance" "chainlaunch_management" {
  count = var.install_chainlaunch_server ? 1 : 0

  ami           = data.aws_ami.ubuntu.id
  instance_type = var.management_instance_type
  key_name      = var.key_pair_name

  subnet_id                   = aws_subnet.fabric_public_subnet.id
  vpc_security_group_ids      = [aws_security_group.chainlaunch_management_sg[0].id]
  associate_public_ip_address = true

  root_block_device {
    volume_type = "gp3"
    volume_size = 50
    encrypted   = true
  }

  user_data = <<-EOF
              #!/bin/bash
              apt-get update
              apt-get install -y docker.io docker-compose curl wget
              systemctl enable docker
              systemctl start docker
              usermod -aG docker ubuntu
              EOF

  tags = {
    Name        = "${var.project_name}-chainlaunch-server"
    Environment = var.environment
    NodeType    = "management"
  }
}

# ==============================================================================
# AWS EC2 INSTANCES - PEER NODES
# ==============================================================================

# Create EC2 instances for peer nodes
resource "aws_instance" "fabric_peers" {
  for_each = var.peer_nodes

  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type
  key_name      = var.key_pair_name

  subnet_id                   = aws_subnet.fabric_public_subnet.id
  vpc_security_group_ids      = [aws_security_group.fabric_peer_sg.id]
  associate_public_ip_address = true

  root_block_device {
    volume_type = "gp3"
    volume_size = 30
    encrypted   = true
  }

  user_data = <<-EOF
              #!/bin/bash
              apt-get update
              apt-get install -y docker.io docker-compose
              systemctl enable docker
              systemctl start docker
              usermod -aG docker ubuntu
              EOF

  tags = {
    Name        = "${var.project_name}-${each.key}"
    Environment = var.environment
    NodeType    = "peer"
    Organization = each.value.organization
  }
}

# ==============================================================================
# AWS EC2 INSTANCES - ORDERER NODES
# ==============================================================================

# Create EC2 instances for orderer nodes
resource "aws_instance" "fabric_orderers" {
  for_each = var.orderer_nodes

  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type
  key_name      = var.key_pair_name

  subnet_id                   = aws_subnet.fabric_public_subnet.id
  vpc_security_group_ids      = [aws_security_group.fabric_orderer_sg.id]
  associate_public_ip_address = true

  root_block_device {
    volume_type = "gp3"
    volume_size = 30
    encrypted   = true
  }

  user_data = <<-EOF
              #!/bin/bash
              apt-get update
              apt-get install -y docker.io docker-compose
              systemctl enable docker
              systemctl start docker
              usermod -aG docker ubuntu
              EOF

  tags = {
    Name        = "${var.project_name}-${each.key}"
    Environment = var.environment
    NodeType    = "orderer"
    Organization = each.value.organization
  }
}

# ==============================================================================
# CHAINLAUNCH RESOURCES - ORGANIZATIONS
# ==============================================================================

# Create Fabric peer organizations
resource "chainlaunch_fabric_organization" "peer_orgs" {
  for_each = toset([for node_key, node in var.peer_nodes : node.organization])

  msp_id      = "${each.key}MSP"
  description = "Peer organization ${each.key}"
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
}

# Create Fabric orderer organizations
resource "chainlaunch_fabric_organization" "orderer_orgs" {
  for_each = toset([for node_key, node in var.orderer_nodes : node.organization])

  msp_id      = "${each.key}MSP"
  description = "Orderer organization ${each.key}"
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
}

# ==============================================================================
# CHAINLAUNCH RESOURCES - PEER NODES
# ==============================================================================

# Create Chainlaunch Fabric peer nodes using AWS instance IPs
resource "chainlaunch_fabric_peer" "peers" {
  for_each = var.peer_nodes

  # Wait for EC2 instance to be ready
  depends_on = [aws_instance.fabric_peers]

  name            = each.key
  organization_id = chainlaunch_fabric_organization.peer_orgs[each.value.organization].id
  msp_id          = "${each.value.organization}MSP"
  mode            = "docker"
  version         = var.fabric_version

  # Use AWS instance public IP for external endpoint
  external_endpoint = "${aws_instance.fabric_peers[each.key].public_ip}:7051"

  # Internal addresses (Docker network)
  listen_address            = "0.0.0.0:7051"
  chaincode_address         = "${aws_instance.fabric_peers[each.key].public_ip}:7052"
  events_address            = "${aws_instance.fabric_peers[each.key].public_ip}:7053"
  operations_listen_address = "0.0.0.0:9443"

  # Map external IPs to internal addresses
  address_overrides = [
    {
      from = "${aws_instance.fabric_peers[each.key].public_ip}:7051"
      to   = "${aws_instance.fabric_peers[each.key].private_ip}:7051"
    }
  ]

  # Domain names
  domain_names = [
    each.key,
    "${each.key}.${each.value.organization}.com",
    "localhost"
  ]

  # Certificate configuration
  certificate_expiration = 365
  auto_renewal_enabled   = true
  auto_renewal_days      = 30

  # Environment variables
  environment = {
    CORE_PEER_GOSSIP_USELEADERELECTION = "true"
    CORE_PEER_GOSSIP_ORGLEADER         = "false"
    CORE_PEER_PROFILE_ENABLED          = "true"
    FABRIC_LOGGING_SPEC                = "INFO"
    CORE_PEER_GOSSIP_EXTERNALENDPOINT  = "${aws_instance.fabric_peers[each.key].public_ip}:7051"
  }
}

# ==============================================================================
# CHAINLAUNCH RESOURCES - ORDERER NODES
# ==============================================================================

# Create Chainlaunch Fabric orderer nodes using AWS instance IPs
resource "chainlaunch_fabric_orderer" "orderers" {
  for_each = var.orderer_nodes

  # Wait for EC2 instance to be ready
  depends_on = [aws_instance.fabric_orderers]

  name            = each.key
  organization_id = chainlaunch_fabric_organization.orderer_orgs[each.value.organization].id
  msp_id          = "${each.value.organization}MSP"
  mode            = "docker"
  version         = var.fabric_version
  consensus       = "etcdraft"

  # Use AWS instance public IP for external endpoint
  external_endpoint = "${aws_instance.fabric_orderers[each.key].public_ip}:7050"

  # Internal addresses (Docker network)
  listen_address            = "0.0.0.0:7050"
  operations_listen_address = "0.0.0.0:9443"

  # Map external IPs to internal addresses
  address_overrides = [
    {
      from = "${aws_instance.fabric_orderers[each.key].public_ip}:7050"
      to   = "${aws_instance.fabric_orderers[each.key].private_ip}:7050"
    }
  ]

  # Domain names
  domain_names = [
    each.key,
    "${each.key}.${each.value.organization}.com",
    "localhost"
  ]

  # Certificate configuration
  certificate_expiration = 365
  auto_renewal_enabled   = true
  auto_renewal_days      = 30

  # Environment variables
  environment = {
    ORDERER_GENERAL_LISTENADDRESS    = "0.0.0.0"
    ORDERER_GENERAL_LISTENPORT       = "7050"
    FABRIC_LOGGING_SPEC              = "INFO"
    ORDERER_OPERATIONS_LISTENADDRESS = "0.0.0.0:9443"
  }
}
