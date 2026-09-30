terraform {
  required_version = ">= 1.5.0, < 2.0.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
  }
}

provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
}

variable "subscription_id" {
  type        = string
  description = "Azure subscription where the demo will run."
}

variable "registry_name" {
  type        = string
  description = "Globally unique ACR name, using 5-50 letters and numbers."
}

variable "location" {
  type        = string
  description = "Azure region for the demo."
  default     = "eastus2"
}

resource "azurerm_resource_group" "demo" {
  name     = "matrix-challenge"
  location = var.location
}

resource "azurerm_container_app_environment" "demo" {
  name                = "matrix-challenge-env"
  resource_group_name = azurerm_resource_group.demo.name
  location            = azurerm_resource_group.demo.location
}

resource "azurerm_container_registry" "demo" {
  name                = var.registry_name
  resource_group_name = azurerm_resource_group.demo.name
  location            = azurerm_resource_group.demo.location
  sku                 = "Basic"
  admin_enabled       = true
}

output "resource_group_id" {
  value = azurerm_resource_group.demo.id
}

output "registry_name" {
  value = azurerm_container_registry.demo.name
}

output "registry_username" {
  value = azurerm_container_registry.demo.admin_username
}

output "registry_password" {
  value     = azurerm_container_registry.demo.admin_password
  sensitive = true
}
