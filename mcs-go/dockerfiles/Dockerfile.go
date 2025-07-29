# Go development environment
FROM codercom/code-server:latest

# Metadata
LABEL description="MCS Go development environment"

# Switch to root for Go installation
USER root

# Install Go with architecture detection
ARG TARGETARCH
ENV GO_VERSION=1.21.5
RUN GOARCH=${TARGETARCH:-amd64} && \
    wget -O go.tar.gz "https://golang.org/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz" && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz

# Set Go environment variables
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/home/coder/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# Create Go workspace directories with proper ownership
RUN mkdir -p /home/coder/go/bin /home/coder/go/src /home/coder/go/pkg && \
    chown -R coder:coder /home/coder/go

# Install Go tools
RUN go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Fix ownership of installed tools
RUN chown -R coder:coder /home/coder/go

# Make Go available system-wide
RUN echo 'export PATH="/usr/local/go/bin:${PATH}"' >> /etc/profile.d/go.sh && \
    echo 'export GOPATH="/home/coder/go"' >> /etc/profile.d/go.sh && \
    echo 'export PATH="${GOPATH}/bin:${PATH}"' >> /etc/profile.d/go.sh && \
    chmod +x /etc/profile.d/go.sh

# Add Go paths to coder user's bashrc
RUN echo '' >> /home/coder/.bashrc && \
    echo '# Go environment' >> /home/coder/.bashrc && \
    echo 'export PATH="/usr/local/go/bin:${PATH}"' >> /home/coder/.bashrc && \
    echo 'export GOPATH="/home/coder/go"' >> /home/coder/.bashrc && \
    echo 'export PATH="${GOPATH}/bin:${PATH}"' >> /home/coder/.bashrc && \
    chown coder:coder /home/coder/.bashrc

# Switch back to coder user
USER coder

# Verify installation
RUN go version