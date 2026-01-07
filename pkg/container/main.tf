terraform {
  required_providers {
    # Keeps your Docker provider from Phase 2
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0"
    }
    # Adds the Azure provider for Phase 3
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

# --- AZURE CONFIGURATION (Phase 3) ---
provider "azurerm" {
  features {}
  # This tells Terraform to work locally while we wait for your credits
  skip_provider_registration = true
}

resource "azurerm_resource_group" "logistics_rg" {
  name     = "muskan-logistics-rg"
  location = "East US"
}

# --- DOCKER CONFIGURATION (Phase 2) ---
provider "docker" {}

resource "docker_image" "logistics_app" {
  name         = "logistics-app:latest"
  keep_locally = true
}

resource "docker_container" "logistics_server" {
  image = docker_image.logistics_app.image_id
  name  = "logistics-container"
  ports {
    internal = 5000
    external = 5000
  }
}