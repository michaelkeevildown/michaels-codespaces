# MCS-Go Test Priority List

## 🚨 Build Issues to Fix First

1. **dockerfiles/Dockerfile.go** - This appears to be a Dockerfile, not Go code. Should be removed from Go compilation.
2. **internal/update/updater.go** - Has duplicate function declarations with checker.go
3. **internal/utils/command.go** - Has unused import

## 📊 Test Files Needed (By Priority)

### Priority 1: Core Entry Points & CLI (No tests exist)
- [ ] `cmd/mcs/main_test.go` - Main application entry
- [ ] `internal/cli/create_test.go` - Core create functionality
- [ ] `internal/cli/list_test.go` - List codespaces
- [ ] `internal/cli/destroy_test.go` - Destroy codespaces
- [ ] `internal/cli/exec_test.go` - Execute commands

### Priority 2: Core Business Logic (No tests exist)
- [ ] `internal/codespace/codespace_test.go` - Codespace struct and methods
- [ ] `internal/codespace/create_test.go` - Creation logic
- [ ] `internal/codespace/manager_test.go` - Management operations
- [ ] `internal/docker/client_test.go` - Docker client operations
- [ ] `internal/docker/compose_test.go` - Docker compose operations

### Priority 3: Utilities (No tests exist)
- [ ] `pkg/utils/names_test.go` - Name generation and collision detection
- [ ] `pkg/utils/repository_test.go` - Git URL parsing
- [ ] `pkg/utils/filesystem_test.go` - File operations
- [ ] `pkg/utils/network_test.go` - Network utilities
- [ ] `pkg/utils/system_test.go` - System information

### Priority 4: Configuration & Components (No tests exist)
- [ ] `internal/config/manager_test.go` - Config management
- [ ] `internal/config/paths_test.go` - Path handling
- [ ] `internal/components/registry_test.go` - Component registration
- [ ] `internal/components/selector_test.go` - Component selection

### Priority 5: Supporting Features (No tests exist)
- [ ] `internal/update/checker_test.go` - Update checking
- [ ] `internal/update/docker_images_test.go` - Image updates
- [ ] `internal/git/clone_test.go` - Git operations
- [ ] `internal/ports/allocator_test.go` - Port allocation
- [ ] `internal/shell/shell_test.go` - Shell operations

### Priority 6: UI Components (No tests exist)
- [ ] `internal/ui/progress_test.go` - Progress indicators
- [ ] `internal/ui/header_test.go` - Header display

### Priority 7: Additional CLI Commands (No tests exist)
- [ ] `internal/cli/backup_test.go`
- [ ] `internal/cli/cleanup_test.go`
- [ ] `internal/cli/doctor_test.go`
- [ ] `internal/cli/info_test.go`
- [ ] `internal/cli/logs_test.go`
- [ ] `internal/cli/recover_test.go`
- [ ] `internal/cli/reset_password_test.go`
- [ ] `internal/cli/setup_test.go`
- [ ] `internal/cli/status_test.go`
- [ ] `internal/cli/update_ip_test.go`
- [ ] `internal/cli/update_images_test.go`
- [ ] `internal/cli/version_test.go`

## 📈 Coverage Targets

### Phase 1 (Immediate)
- Fix build errors
- Add tests for main.go and core CLI commands
- Target: 20% coverage

### Phase 2 (Short-term) 
- Test core business logic (codespace, docker)
- Test all utilities in pkg/utils
- Target: 40% coverage

### Phase 3 (Medium-term)
- Test all CLI commands
- Test configuration and components
- Target: 60% coverage

### Phase 4 (Long-term)
- Add integration tests
- Test error scenarios
- Target: 80%+ coverage

## 🧪 Testing Approach

1. **Unit Tests**: Mock external dependencies (Docker, filesystem)
2. **Table-Driven Tests**: Use for utilities and validation logic
3. **Integration Tests**: Test actual Docker operations in CI
4. **CLI Tests**: Use cobra testing utilities
5. **Error Cases**: Test error handling paths explicitly

---
Generated: 2025-07-29T07:50:00Z