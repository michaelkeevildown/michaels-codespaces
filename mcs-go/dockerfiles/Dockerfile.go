# Go development environment
FROM mcs/code-server-base:latest

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

# Install Go tools
RUN go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Switch back to coder user
USER coder

# Set up Go workspace
RUN mkdir -p ~/go/{bin,src,pkg}

# Verify installation
RUN go version