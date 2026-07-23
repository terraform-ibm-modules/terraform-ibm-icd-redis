# Databases for Redis - Fully Configurable (gen2) Deployable Architecture

This deployable architecture deploys and configures an IBM Cloud [Databases for Redis](https://www.ibm.com/products/databases-for-redis) instance using the **gen2 hosting model**, which provisions instances on dedicated compute hosts.

## Overview

The gen2 hosting model uses dedicated, isolated compute resources for your Redis deployment, providing improved performance predictability and resource isolation compared to the multitenant hosting model. You select a specific host flavor (e.g., `b3c.4x16`) that defines the compute resources available per member.

## Supported host flavors (`member_host_flavor`)

| Flavor | vCPU | Memory |
|--------|------|--------|
| `b3c.4x16` | 4 | 16 GB |
| `b3c.8x32` | 8 | 32 GB |
| `b3c.16x64` | 16 | 64 GB |
| `b3c.32x128` | 32 | 128 GB |
| `m3c.8x64` | 8 | 64 GB |
| `m3c.16x128` | 16 | 128 GB |
| `m3c.30x240` | 30 | 240 GB |

## Usage

For information about configuring complex input variables (autoscaling, configuration, users, service credentials), see [DA-types.md](./DA-types.md).

## Features

- Deploys a Databases for Redis instance using gen2 dedicated compute hosts
- Optional KMS encryption (Key Protect or Hyper Protect Crypto Services)
- Optional service credentials stored in Secrets Manager
- Optional autoscaling configuration
- Optional Redis database configuration tuning
- Support for using an existing Redis instance (read-only mode)

## Required IAM permissions

| Service | Role | Notes |
|---------|------|-------|
| Resource group | Viewer | Required in the resource group you want to provision in. |
| `databases-for-redis` | Editor | Required to configure an instance of Databases for Redis. |
| `kms` | Editor | [Optional] Required to create keys when using Key Protect for encryption. |
| `hs-crypto` | Editor | [Optional] Required to create keys when using HPCS for encryption. |
