# Multi-stage Containerfile for OpenNotes
# Base stage: mise + build dependencies + isolated test environment
# Default stage: development environment

# =============================================================================
# Stage: base
# Base environment with mise and build dependencies
# =============================================================================
FROM docker.io/jdxcode/mise:latest AS base

# Install build dependencies for DuckDB
USER root
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*;

# Create isolated test user with clean home directory
RUN useradd -m -d /home/testuser -s /bin/bash testuser;

# Create expected config directories (empty, isolated)
RUN mkdir -p /home/testuser/.config/opennotes \
    && mkdir -p /home/testuser/.cache \
    && mkdir -p /home/testuser/.local/share/mise \
    && chown -R testuser:testuser /home/testuser;

# Create entrypoint script inline
RUN echo '#!/bin/bash\n\
set -euo pipefail\n\
\n\
# Trust mise config if present\n\
if [ -f mise.toml ] || [ -f .mise.toml ]; then\n\
    echo "Trusting mise configuration..."\n\
    mise trust 2>/dev/null || true\n\
fi\n\
\n\
# Install mise tools if config exists\n\
if [ -f mise.toml ] || [ -f .mise.toml ]; then\n\
    echo "Installing mise tools..."\n\
    mise install || true\n\
fi\n\
\n\
# Execute provided command\n\
exec "$@"\n\
' > /usr/local/bin/entrypoint.sh && chmod +x /usr/local/bin/entrypoint.sh;

# App directory (will be mounted)
RUN mkdir -p /app && chown testuser:testuser /app;

USER testuser
WORKDIR /app

# Set mise environment variables for isolated operation
ENV HOME=/home/testuser
ENV MISE_YES=1
ENV MISE_DATA_DIR="/home/testuser/.local/share/mise"
ENV MISE_CACHE_DIR="/home/testuser/.cache/mise"
ENV PATH="/home/testuser/.local/share/mise/shims:/usr/local/bin:$PATH"

# =============================================================================
# Stage: default (final)
# Development environment - trust and install tools on startup
# =============================================================================
FROM base AS default

USER testuser
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
