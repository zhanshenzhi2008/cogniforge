# Cogniforge Knowledge Service

FastAPI-based knowledge processing service for RAG pipelines.

## Features

- **Multi-format Document Parsing**: PDF, DOCX, TXT, MD, HTML
- **Intelligent Text Chunking**: Recursive character and sentence splitters
- **Flexible Embeddings**: OpenAI API or local sentence transformers
- **Vector Storage**: PostgreSQL with pgvector extension
- **Semantic Search**: Cosine similarity search with filtering

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    FastAPI Server (8085)                │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │  Parsers    │  │  Splitters   │  │  Embedding   │   │
│  │  PDF        │  │  Recursive   │  │  OpenAI     │   │
│  │  DOCX       │→ │  Sentence    │→ │  Local      │   │
│  │  TXT/MD     │  │              │  │              │   │
│  │  HTML       │  │              │  │              │   │
│  └─────────────┘  └─────────────┘  └─────────────┘   │
│          │                │                │              │
│          └────────────────┼────────────────┘              │
│                           ▼                               │
│  ┌─────────────────────────────────────────────────┐    │
│  │              Vector Store (pgvector)              │    │
│  │              - HNSW Index for fast search        │    │
│  └─────────────────────────────────────────────────┘    │
│                           │                               │
└───────────────────────────┼───────────────────────────────┘
                            ▼
              ┌─────────────────────────────┐
              │    PostgreSQL (5432/5433)    │
              │    - Documents table         │
              │    - Vectors table          │
              └─────────────────────────────┘
```

## Quick Start

### 1. Install Dependencies

```bash
cd llm/knowledge
pip install -r requirements.txt
```

### 2. Configure Environment

Create a `.env` file or set environment variables:

```bash
# Database
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5433
export POSTGRES_DB=cogniforge
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=your_password

# Embedding (choose one)
export EMBEDDER_TYPE=openai  # or "local"
export OPENAI_API_KEY=sk-...  # Required for OpenAI

# Optional
export CHUNK_SIZE=512
export CHUNK_OVERLAP=50
```

### 3. Enable pgvector Extension

```sql
-- In PostgreSQL
CREATE EXTENSION IF NOT EXISTS vector;

-- Verify
SELECT extname FROM pg_extension WHERE extname = 'vector';
```

### 4. Start the Server

```bash
uvicorn app.main:app --host 0.0.0.0 --port 8085 --reload
```

## API Endpoints

### Health Check

```bash
curl http://localhost:8085/health
```

### Process Document

```bash
curl -X POST http://localhost:8085/api/knowledge/process \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/document.pdf",
    "document_id": "doc_123",
    "collection_name": "my_knowledge_base"
  }'
```

### Search

```bash
curl -X POST http://localhost:8085/api/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is machine learning?",
    "collection_name": "my_knowledge_base",
    "top_k": 5,
    "min_score": 0.5
  }'
```

### Upload & Process

```bash
curl -X POST http://localhost:8085/api/knowledge/upload \
  -F "file=@document.pdf" \
  -F "document_id=doc_456" \
  -F "collection_name=my_knowledge_base"
```

## Directory Structure

```
llm/knowledge/
├── app/
│   ├── __init__.py
│   └── main.py              # FastAPI application
├── parsers/                  # Document parsers
│   ├── __init__.py
│   ├── base.py
│   ├── pdf_parser.py
│   ├── docx_parser.py
│   ├── txt_parser.py
│   └── html_parser.py
├── splitters/               # Text chunkers
│   ├── __init__.py
│   ├── base.py
│   └── recursive_splitter.py
├── embedding/               # Embedding models
│   ├── __init__.py
│   ├── base.py
│   ├── openai_embedder.py
│   └── local_embedder.py
├── vector_store/            # Vector databases
│   ├── __init__.py
│   ├── base.py
│   └── pgvector_store.py
├── services/                # Business logic
│   ├── __init__.py
│   └── document_processor.py
├── utils/                   # Utilities
├── models/                 # Data models
├── requirements.txt        # Dependencies
└── README.md
```

## Development

### Run Tests

```bash
# Install dev dependencies
pip install pytest pytest-asyncio httpx

# Run tests
pytest tests/
```

### API Documentation

Once running, visit:
- Swagger UI: http://localhost:8085/docs
- ReDoc: http://localhost:8085/redoc

## License

MIT
