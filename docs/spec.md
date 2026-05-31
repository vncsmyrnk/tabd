# Service generator for LUKS containers

## 1. Objective

Define a configuration generator for service managers to mount LUKS encrypted devices and optionally stow their contents.

## 2. Core Principles

- **Declarative Configuration**: Users define the desired state of encrypted volumes.
- **Service-Based Lifecycle**: Leverage existing service managers (e.g., systemd) for robust lifecycle management (mounting, unmounting, dependency handling).
- **Non-Invasive Stowing**: Support symbolic linking (stowing) of decrypted data into user-defined locations (e.g., `$HOME`) to separate storage from presentation.
- **Portability**: The generator should produce artifacts that can be deployed across different systems with minimal dependencies.

## 3. Terminology

- **Container**: A LUKS-encrypted file or block device.
- **Mapper**: The decrypted virtual device exposed via Device Mapper (e.g., `/dev/mapper/my-vault`).
- **Mountpoint**: The directory where the decrypted container is attached.
- **Stow Target**: The directory where the contents of the Mountpoint are symlinked.
- **Generator**: The tool that translates configuration into service units.

## 4. Functional Requirements

### 4.1. Lifecycle Management
The tool MUST support:
- Opening a LUKS container (`cryptsetup open`).
- Closing a LUKS container (`cryptsetup close`).
- Mounting the filesystem within the container.
- Unmounting the filesystem.
- Generating service units (e.g., systemd `.service` or `.mount` units) to automate these steps.

### 4.2. Stowing Logic
If "stow" is enabled:
- The tool MUST symlink entries from the Mountpoint to the Stow Target.
- The tool SHOULD handle conflicts (e.g., pre-existing files in the Stow Target).
- The tool MUST remove symlinks upon unmounting.

## 5. Proposed Architecture

### 5.1. Input
For service manager configuration file generator: container image path, mountpoint and stow target (optional).
For managing the actual services: listing, starting and stopping CLI commands.

### 5.2. Output
A service manager (e.g. systemd) configuration file content.
