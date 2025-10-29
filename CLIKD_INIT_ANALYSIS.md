# Clikd Init Command - Comprehensive Analysis & Design

## Executive Summary

The `clikd init` command will establish a complete local development environment for Clikd's gaming social platform backend, following the Supabase CLI principle where frontend developers can execute one command to get a fully functional local backend.

## Analysis Findings

### Clikd Backend Architecture

**Core Services:**
- **Gate Service (Port 3001)**: Authentication service handling OAuth, OIDC, session management
- **API Service (Port 3002)**: Main REST API with comprehensive gaming social features
- **Realtime Service (Port 3003)**: WebSocket connections for live chat, notifications, presence
- **Media Service (Port 3004)**: File upload, processing, CDN integration

**Infrastructure Stack:**
- **PostgreSQL (5432)**: Primary database for auth and main schemas
- **ScyllaDB (9042)**: High-performance analytics and time-series data
- **KeyDB (6379)**: Redis-compatible caching and session storage
- **MinIO (9000/9001)**: S3-compatible object storage
- **NATS (4222/8222/6222)**: Message queuing and event streaming
- **APISIX (9080/9092/9443)**: API Gateway with rate limiting and security

**Security Architecture:**
- Internal port isolation (3001-3004 not externally exposed)
- APISIX as single external entry point
- Branch-isolated environments with sanitized naming
- GHCR container registry at `ghcr.io/clikd-inc/*`

### Supabase CLI Pattern Analysis

**Key Learnings:**
- Direct Docker API usage (Bollard) instead of docker-compose
- Health check-based service orchestration with retry logic
- Template-driven configuration with environment substitution
- Service discovery via Docker networks
- Graceful startup/shutdown sequences respecting dependencies
- Real-time status monitoring and logging

### Current Development Gaps

**Missing Components:**
- No local development orchestration
- No branch-based environment isolation
- No automated service health monitoring
- No configuration templating system
- No developer onboarding workflow

## Init Command Design

### Primary Objectives

1. **One-Command Setup**: `clikd init` creates complete local backend
2. **Branch Isolation**: Each git branch gets isolated databases and services
3. **Health Monitoring**: Services start in dependency order with health checks
4. **Configuration Management**: Template-based configs with branch substitution
5. **Developer Experience**: Clear feedback, progress indicators, error recovery

### Technical Implementation

#### Docker Network Strategy
```
Network: clikd-{branch}
- Services discover each other by name
- Isolated per git branch
- Automatic cleanup on environment destruction
```

#### Service Startup Sequence
```
1. Infrastructure Layer:
   - PostgreSQL with branch databases
   - ScyllaDB with branch keyspaces
   - KeyDB with branch prefixes
   - MinIO with branch buckets
   - NATS with branch subjects

2. Backend Services:
   - Gate (waits for databases)
   - API (waits for Gate + databases)
   - Realtime (waits for API + NATS)
   - Media (waits for MinIO + API)

3. Gateway Layer:
   - APISIX (waits for all backend services)
```

#### Configuration Templating
```
Templates in ~/.clikd/templates/:
- docker-compose.{branch}.yml
- apisix-config.{branch}.yml
- gate.{branch}.toml
- .env.{branch}

Variables:
- {{BRANCH}}: Sanitized git branch name
- {{POSTGRES_DB_AUTH}}: clikd_auth_{branch}
- {{POSTGRES_DB_MAIN}}: clikd_rig_{branch}
- {{SCYLLA_KEYSPACE}}: clikd_{branch}
- {{KEYDB_PREFIX}}: clikd_{branch}
- {{CONTAINER_PREFIX}}: clikd-{branch}
```

#### Image Strategy
```
Main branch: ghcr.io/clikd-inc/{service}:latest
Feature branches: ghcr.io/clikd-inc/{service}:{sanitized-branch}

Fallback: Local build if image not found
Health checks: Service-specific endpoints
Logs: Aggregated with service prefixes
```

### Command Workflow

#### `clikd init` Execution Flow

1. **Environment Detection**
   - Detect git branch and repository status
   - Check Docker availability and permissions
   - Verify GHCR access for private repositories

2. **Configuration Generation**
   - Create `~/clikd/` directory structure
   - Create config.toml
   - Generate branch-specific configuration files
   - Template substitution with branch variables

3. **Docker Network Setup**
   - Create isolated network: `clikd-{branch}`
   - Configure DNS resolution between services
   - Set up volume mounts for persistence

4. **Infrastructure Startup**
   - PostgreSQL with auto-created databases
   - ScyllaDB with keyspace initialization
   - KeyDB with branch-prefixed keys
   - MinIO with bucket creation
   - NATS with subject configuration

5. **Service Orchestration**
   - Pull/build container images for branch
   - Start services in dependency order
   - Wait for health checks before proceeding
   - Configure service discovery

6. **Gateway Configuration**
   - Generate APISIX routes for branch
   - Configure rate limiting and security
   - Set up SSL termination if needed

7. **Verification & Output**
   - Verify all services healthy
   - Display service URLs and status
   - Generate `.env.local` for frontend
   - Show next steps and documentation

### Developer Experience

#### Interactive TUI Features
- Real-time service status dashboard
- Progress indicators with health checks
- Expandable logs for troubleshooting
- Service restart/rebuild capabilities
- Environment cleanup options

#### Configuration Output
```
Frontend developers receive:
- API_URL=http://localhost:9080/api
- REALTIME_URL=ws://localhost:9080/realtime
- MEDIA_URL=http://localhost:9080/media
- Branch-specific database connections
```

#### Error Recovery
- Automatic retry with exponential backoff
- Detailed error messages with solutions
- Cleanup on failure with state preservation
- Resume capability for partial failures

### Integration Points

#### Git Integration
- Automatic branch detection
- Hook into git checkout for environment switching
- Cleanup of abandoned branch environments
- Status indication of environment sync

#### CI/CD Integration
- Skip interactive features in CI
- JSON output mode for automation
- Environment validation commands
- Cleanup commands for CI runners

## Implementation Phases

### Phase 1: Core Infrastructure
- Bollard Docker client integration
- Configuration templating system
- Basic service orchestration
- Health check framework

### Phase 2: Service Management
- Complete service definitions
- Dependency resolution
- Health monitoring
- Log aggregation

### Phase 3: Developer Experience
- Interactive TUI dashboard
- Error recovery systems
- Documentation generation
- Performance optimization

### Phase 4: Advanced Features
- Multi-environment support
- Service scaling capabilities
- Performance monitoring
- Integration testing

## Success Metrics

1. **Setup Time**: Complete environment in under 3 minutes
2. **Reliability**: 99% success rate on clean systems
3. **Developer Adoption**: Primary onboarding method
4. **Maintenance**: Self-healing and auto-updating
5. **Documentation**: Generated and always current

## Conclusion

The `clikd init` command will transform Clikd's developer experience by providing instant, isolated, branch-specific backend environments. Following Supabase's proven patterns while adapting to Clikd's unique gaming social platform requirements ensures both reliability and developer satisfaction.

This design enables true frontend-first development where backend complexity is completely abstracted away, allowing developers to focus on building exceptional user experiences.
