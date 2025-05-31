#!/bin/bash

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
SKIP_TESTS=false
SKIP_LINT=false
DEV_MODE=false
RACE_DETECTOR=false
DEBUG_BUILD=false
VERBOSE=false

# Print step description
print_step() {
    echo -e "${YELLOW}==>${NC} $1"
}

# Print usage
usage() {
    echo -e "${BLUE}Usage:${NC} $0 [options]"
    echo
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -d, --dev           Development mode (faster builds, no optimization)"
    echo "  --skip-tests        Skip running tests"
    echo "  --skip-lint         Skip running linter"
    echo "  --race             Enable race condition detector"
    echo "  --debug            Include debug information in binary"
    echo "  -v, --verbose      Verbose output"
    echo "  --clean-only       Only clean build artifacts"
    echo "  --test-only        Only run tests"
    echo "  --build-only       Only build for current platform"
    echo "  --dist-only        Only create distribution packages"
    echo
    echo "Environment variables:"
    echo "  EXTRA_LDFLAGS      Additional ldflags for go build"
    echo "  EXTRA_TAGS         Additional build tags"
    echo "  PLATFORMS          Override default build platforms (format: 'os/arch os/arch ...')"
}

# Parse command line arguments
parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            -h|--help)
                usage
                exit 0
                ;;
            -d|--dev)
                DEV_MODE=true
                ;;
            --skip-tests)
                SKIP_TESTS=true
                ;;
            --skip-lint)
                SKIP_LINT=true
                ;;
            --race)
                RACE_DETECTOR=true
                ;;
            --debug)
                DEBUG_BUILD=true
                ;;
            -v|--verbose)
                VERBOSE=true
                ;;
            --clean-only)
                clean
                exit 0
                ;;
            --test-only)
                run_tests
                exit 0
                ;;
            --build-only)
                build
                exit 0
                ;;
            --dist-only)
                create_dist
                exit 0
                ;;
            *)
                echo -e "${RED}Error: Unknown option $1${NC}"
                usage
                exit 1
                ;;
        esac
        shift
    done
}

# Check if required tools are installed
check_requirements() {
    print_step "Checking build requirements"
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: go is not installed${NC}"
        exit 1
    fi

    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "${GO_VERSION}" < "1.16" ]]; then
        echo -e "${RED}Error: Go version must be 1.16 or higher (current: ${GO_VERSION})${NC}"
        exit 1
    fi

    # Check for required build tools
    if [ "$VERBOSE" = true ]; then
        echo "Go version: $GO_VERSION"
        go env
    fi

    echo -e "${GREEN}✓ All requirements satisfied${NC}"
}

# Clean build artifacts
clean() {
    print_step "Cleaning build artifacts"
    rm -rf bin/
    rm -rf dist/
    go clean -cache -testcache
    echo -e "${GREEN}✓ Clean completed${NC}"
}

# Run tests
run_tests() {
    if [ "$SKIP_TESTS" = true ]; then
        echo -e "${YELLOW}Skipping tests${NC}"
        return
    fi

    print_step "Running tests"
    
    TEST_FLAGS="-v"
    if [ "$RACE_DETECTOR" = true ]; then
        TEST_FLAGS="$TEST_FLAGS -race"
    fi
    if [ "$VERBOSE" = true ]; then
        TEST_FLAGS="$TEST_FLAGS -json"
    fi

    go test $TEST_FLAGS ./...
    echo -e "${GREEN}✓ Tests passed${NC}"
}

# Run linter
run_lint() {
    if [ "$SKIP_LINT" = true ]; then
        echo -e "${YELLOW}Skipping linter${NC}"
        return
    fi

    print_step "Running linter"
    if ! command -v golangci-lint &> /dev/null; then
        echo -e "${YELLOW}Warning: golangci-lint not found, skipping linting${NC}"
        return
    fi
    
    LINT_FLAGS=""
    if [ "$VERBOSE" = true ]; then
        LINT_FLAGS="--verbose"
    fi
    
    golangci-lint run $LINT_FLAGS
    echo -e "${GREEN}✓ Lint passed${NC}"
}

# Build binary
build() {
    print_step "Building superclass"
    
    # Create build directory
    mkdir -p bin

    # Prepare build flags
    BUILD_FLAGS=""
    LDFLAGS="-s -w" # Default: stripped, no debug info
    
    if [ "$DEV_MODE" = true ]; then
        LDFLAGS="" # No stripping in dev mode
    fi
    
    if [ "$DEBUG_BUILD" = true ]; then
        LDFLAGS="" # No stripping for debug builds
    fi
    
    if [ "$RACE_DETECTOR" = true ]; then
        BUILD_FLAGS="$BUILD_FLAGS -race"
    fi
    
    # Add any extra build tags
    if [ -n "$EXTRA_TAGS" ]; then
        BUILD_FLAGS="$BUILD_FLAGS -tags '$EXTRA_TAGS'"
    fi
    
    # Add any extra ldflags
    if [ -n "$EXTRA_LDFLAGS" ]; then
        LDFLAGS="$LDFLAGS $EXTRA_LDFLAGS"
    fi

    # Build for the current platform
    GOOS=$(go env GOOS)
    GOARCH=$(go env GOARCH)
    OUTPUT="bin/superclass"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="$OUTPUT.exe"
    fi

    if [ "$VERBOSE" = true ]; then
        echo "Building with:"
        echo "  GOOS=$GOOS"
        echo "  GOARCH=$GOARCH"
        echo "  Flags=$BUILD_FLAGS"
        echo "  Ldflags=$LDFLAGS"
    fi

    go build $BUILD_FLAGS -ldflags="$LDFLAGS" -o "$OUTPUT" .
    echo -e "${GREEN}✓ Build completed: $OUTPUT${NC}"

    # Show binary size and info
    ls -lh "$OUTPUT"
    if command -v file &> /dev/null; then
        file "$OUTPUT"
    fi
}

# Create distribution packages
create_dist() {
    print_step "Creating distribution packages"
    
    mkdir -p dist
    VERSION=$(git describe --tags --always || echo "dev")
    
    # Use custom platforms if specified
    if [ -z "$PLATFORMS" ]; then
        PLATFORMS=("linux/amd64" "darwin/amd64" "darwin/arm64" "windows/amd64")
    else
        # Convert space-separated string to array
        read -r -a PLATFORMS <<< "$PLATFORMS"
    fi
    
    if [ "$VERBOSE" = true ]; then
        echo "Creating packages for: ${PLATFORMS[*]}"
    fi
    
    for PLATFORM in "${PLATFORMS[@]}"; do
        # Split PLATFORM into OS and ARCH
        IFS='/' read -r -a array <<< "$PLATFORM"
        GOOS="${array[0]}"
        GOARCH="${array[1]}"
        
        # Set output name
        OUTPUT_NAME="superclass_${VERSION}_${GOOS}_${GOARCH}"
        if [ "$GOOS" = "windows" ]; then
            OUTPUT_NAME="$OUTPUT_NAME.exe"
        fi
        
        echo "Building for $GOOS/$GOARCH..."
        GOOS=$GOOS GOARCH=$GOARCH go build -o "dist/$OUTPUT_NAME" -ldflags="-s -w" .
        
        # Create archive
        pushd dist > /dev/null
        if [ "$GOOS" = "windows" ]; then
            zip "${OUTPUT_NAME%.exe}.zip" "$OUTPUT_NAME" README.md
            rm "$OUTPUT_NAME"
        else
            tar czf "${OUTPUT_NAME}.tar.gz" "$OUTPUT_NAME" README.md
            rm "$OUTPUT_NAME"
        fi
        popd > /dev/null
        
        echo -e "${GREEN}✓ Created package for $GOOS/$GOARCH${NC}"
    done
    
    echo -e "${GREEN}✓ Distribution packages created in dist/${NC}"
    ls -lh dist/
}

# Main build process
main() {
    parse_args "$@"
    check_requirements
    clean
    run_lint
    run_tests
    build
    if [ "$DEV_MODE" = false ]; then
        create_dist
    fi
    echo -e "\n${GREEN}Build process completed successfully!${NC}"
}

# Run main if script is executed directly
if [ "${BASH_SOURCE[0]}" = "$0" ]; then
    main "$@"
fi 