# Superclass
<img width="339" alt="Screenshot 2025-05-31 at 02 00 42" src="https://github.com/user-attachments/assets/5e4ed7d5-082f-4bcd-aad2-fa7d6755ce5b" />

Superclass is a powerful document analysis tool that combines advanced text extraction with AI-powered classification. It supports multiple document formats and provides both a CLI and HTTP server interface.

## Features

### Document Support
- PDF documents
- Microsoft Office (DOCX, XLSX, PPTX)
- OpenDocument (ODT)
- Images (with OCR)
- SVG files (with text extraction)
- HTML files
- Markdown files
- EPUB ebooks
- RTF documents
- Plain text files

### AI Classification
- Multiple AI providers supported:
  - OpenAI (GPT-4, GPT-3.5)
  - Anthropic (Claude)
  - Azure OpenAI
- Classification features:
  - Category detection
  - Predefined categories support
  - Confidence scoring
  - Content summarization
  - Keyword extraction
- Model comparison capabilities
- Advanced feature extraction:
  - Basic statistics (word count, character count, etc.)
  - Language metrics (readability, technicality, formality)
  - Named entity recognition
  - Document structure analysis
  - Sentiment analysis
  - Content complexity assessment
  - Vocabulary richness analysis

### Deployment Options
- Command-line interface
- HTTP server mode
- Docker support

## Installation

### Using Docker

The image is available on GitHub Container Registry:

```bash
# Basic usage
docker pull ghcr.io/adaptive-scale/superclass:latest

# Run with minimal configuration
docker run -p 8083:8083 \
  -e OPENAI_API_KEY=your_openai_key \
  ghcr.io/adaptive-scale/superclass:latest

# Run with common configuration
docker run -p 8083:8083 \
  -e PORT=8083 \
  -e LOG_LEVEL=debug \
  -e MODEL_TYPE=gpt-4 \
  -e MODEL_PROVIDER=openai \
  -e MAX_COST=0.1 \
  -e MAX_LATENCY=30 \
  -e OPENAI_API_KEY=your_openai_key \
  -v /path/to/local/uploads:/tmp/superclass-uploads \
  ghcr.io/adaptive-scale/superclass:latest

# Using environment file
docker run -p 8083:8083 \
  --env-file .env \
  ghcr.io/adaptive-scale/superclass:latest
```

For all available environment variables and their descriptions, see the [Configuration](#configuration) section.

Supported architectures:
- linux/amd64 (x86_64)
- linux/arm64 (Apple Silicon, AWS Graviton)

### Building from Source

#### Prerequisites
- Go 1.19 or later
- Docker with buildx support (for multi-arch builds)
- Make
- Tesseract OCR (for image support)

#### Using Make

```bash
# Build local binary
make build

# Run tests
make test

# Build and push multi-arch Docker image
export GITHUB_TOKEN=your_github_token
export GITHUB_USER=your_github_username
make docker-login
make docker-buildx

# Create a release
VERSION=v1.0.0 make release
```

Available make targets:
```bash
make help  # Show all available targets
```

Common targets:
- `make build`: Build local binary
- `make test`: Run tests
- `make docker-build`: Build Docker image for local architecture
- `make docker-buildx`: Build and push multi-arch Docker images
- `make release VERSION=v1.0.0`: Create and push a new release

Environment variables:
- `REGISTRY`: Container registry (default: ghcr.io)
- `REPOSITORY`: Image repository (default: adaptive-scale/superclass)
- `TAG`: Image tag (default: latest)
- `PLATFORMS`: Target platforms (default: linux/amd64,linux/arm64)
- `GITHUB_TOKEN`: GitHub personal access token
- `GITHUB_USER`: GitHub username

## Usage

### API Endpoints

#### POST /classify
Classify a document:
```bash
# Basic classification
curl -X POST -F "file=@/path/to/document.pdf" http://localhost:8083/classify

# Classification with feature extraction
curl -X POST -F "file=@/path/to/document.pdf" -F "extract_features=true" http://localhost:8080/classify
```

Response with features:
```json
{
  "category": "Technical Documentation",
  "confidence": 0.95,
  "summary": "This document describes...",
  "keywords": ["keyword1", "keyword2"],
  "features": {
    "word_count": 1250,
    "char_count": 6800,
    "sentence_count": 85,
    "avg_word_length": 5.4,
    "unique_word_count": 450,
    "paragraph_count": 25,
    "top_keywords": ["api", "documentation", "endpoints"],
    "named_entities": [
      {"text": "OpenAI", "label": "ORGANIZATION"},
      {"text": "GPT-4", "label": "PRODUCT"}
    ],
    "sentiment_score": 0.2,
    "language_metrics": {
      "readability_score": 65.5,
      "technicality_score": 0.8,
      "formality_score": 0.7,
      "vocabulary_richness": 0.65
    },
    "content_structure": {
      "heading_count": 12,
      "list_count": 8,
      "table_count": 2,
      "code_block_count": 5,
      "image_count": 3,
      "heading_hierarchy": [
        "Introduction",
        "API Reference",
        "Authentication"
      ]
    }
  },
  "raw_text": "Optional extracted text..."
}
```

#### GET /health
Health check endpoint:
```bash
curl http://localhost:8083/health
```

#### POST /features
Extract detailed features from a document without classification:
```bash
# Basic feature extraction
curl -X POST -F "file=@/path/to/document.pdf" http://localhost:8080/features

# Feature extraction with specific model
curl -X POST \
  -F "file=@/path/to/document.pdf" \
  -F "model_provider=anthropic" \
  -F "model_type=claude-3-opus" \
  http://localhost:8080/features
```

Response:
```json
{
  "word_count": 1250,
  "char_count": 6800,
  "sentence_count": 85,
  "avg_word_length": 5.4,
  "unique_word_count": 450,
  "paragraph_count": 25,
  "top_keywords": ["api", "documentation", "endpoints"],
  "named_entities": [
    {"text": "OpenAI", "label": "ORGANIZATION"},
    {"text": "GPT-4", "label": "PRODUCT"}
  ],
  "sentiment_score": 0.2,
  "language_metrics": {
    "readability_score": 65.5,
    "technicality_score": 0.8,
    "formality_score": 0.7,
    "vocabulary_richness": 0.65
  },
  "content_structure": {
    "heading_count": 12,
    "list_count": 8,
    "table_count": 2,
    "code_block_count": 5,
    "image_count": 3,
    "heading_hierarchy": [
      "Introduction",
      "API Reference",
      "Authentication"
    ]
  }
}
```

Parameters:
- `file`: The document file to analyze (required)
- `model_provider`: AI provider to use (optional, defaults to environment setting)
- `model_type`: Specific model to use (optional, defaults to environment setting)
- `raw_text`: Include extracted text in response (optional, default: false)

## Configuration

### Environment Variables

#### Server Configuration
- `PORT`: Server port (default: 8083)
- `UPLOAD_DIR`: Directory for temporary file uploads (default: /tmp/superclass-uploads)
- `LOG_LEVEL`: Logging level (default: debug)

#### Model Configuration
- `MODEL_TYPE`: AI model to use (default: gpt-4)
- `MODEL_PROVIDER`: AI provider to use (default: openai)
- `MAX_COST`: Maximum cost per request (default: 0.1)
- `MAX_LATENCY`: Maximum latency in seconds (default: 30)
- `EXTRACT_FEATURES`: Enable feature extraction by default (default: false)
- `FEATURE_MODEL`: Model to use for feature extraction (default: same as MODEL_TYPE)

#### Classification Configuration
- `PREDEFINED_CATEGORIES`: Comma-separated list of allowed categories (e.g., "Technology,Business,Science")
- `ENFORCE_CATEGORIES`: Whether to strictly enforce predefined categories (default: false)

#### API Keys
- `OPENAI_API_KEY`: OpenAI API key for GPT models
- `ANTHROPIC_API_KEY`: Anthropic API key for Claude models
- `AZURE_OPENAI_API_KEY`: Azure OpenAI API key for Azure deployments

#### Build & Deployment
- `REGISTRY`: Container registry (default: ghcr.io)
- `REPOSITORY`: Image repository (default: adaptive-scale/superclass)
- `TAG`: Image tag (default: latest)
- `PLATFORMS`: Target platforms for multi-arch builds (default: linux/amd64,linux/arm64)
- `GITHUB_TOKEN`: GitHub personal access token for GHCR authentication
- `GITHUB_USER`: GitHub username for GHCR authentication
- `VERSION`: Version tag for releases (e.g., v1.0.0)

Example `.env` file:
```env
# Server Configuration
PORT=8083
LOG_LEVEL=debug
UPLOAD_DIR=/tmp/superclass-uploads

# Model Configuration
MODEL_TYPE=gpt-4
MODEL_PROVIDER=openai
MAX_COST=0.1
MAX_LATENCY=30

# Classification Configuration
PREDEFINED_CATEGORIES=Technology,Business,Science
ENFORCE_CATEGORIES=true

# API Keys
OPENAI_API_KEY=your_openai_key
# ANTHROPIC_API_KEY=your_anthropic_key
# AZURE_OPENAI_API_KEY=your_azure_key
```

Example Docker Compose environment:
```yaml
services:
  superclass:
    environment:
      # Server Configuration
      - PORT=8083
      - LOG_LEVEL=debug
      
      # Model Configuration
      - MODEL_TYPE=gpt-4
      - MODEL_PROVIDER=openai
      - MAX_COST=0.1
      - MAX_LATENCY=30
      
      # Classification Configuration
      - PREDEFINED_CATEGORIES=Technology,Business,Science
      - ENFORCE_CATEGORIES=true
      
      # API Keys
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      # - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      # - AZURE_OPENAI_API_KEY=${AZURE_OPENAI_API_KEY}
```

### Model Configuration
Available models:
- OpenAI:
  - gpt-4
  - gpt-4-turbo
  - gpt-3.5-turbo
- Anthropic:
  - claude-3-opus
  - claude-3-sonnet
  - claude-3-haiku
- Azure OpenAI: (depends on deployment)

### Classification Categories

When using predefined categories:
1. Set `PREDEFINED_CATEGORIES` to a comma-separated list of categories
2. Optionally set `ENFORCE_CATEGORIES=true` to ensure only predefined categories are returned
3. Categories can also be specified per-request in the API call

Example using predefined categories:
```bash
# Using environment variables
export PREDEFINED_CATEGORIES="Technology,Business,Science,Health,Entertainment"
export ENFORCE_CATEGORIES=true
docker-compose up

# Or in docker-compose.yml
services:
  superclass:
    environment:
      - PREDEFINED_CATEGORIES=Technology,Business,Science
      - ENFORCE_CATEGORIES=true
```

## Development

### Prerequisites
- Go 1.19 or later
- Tesseract OCR (for image support)
- Required dependencies:
  ```bash
  go mod download
  ```

### Building
```bash
./build.sh
```

Build options:
- `--dev`: Development build
- `--race`: Enable race condition detection
- `--debug`: Include debug information

### Testing
```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [UniDoc](https://github.com/unidoc/unioffice) for document processing
- [Tesseract](https://github.com/tesseract-ocr/tesseract) for OCR capabilities
- OpenAI and Anthropic for AI models 
