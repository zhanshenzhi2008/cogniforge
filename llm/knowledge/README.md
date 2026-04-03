# LLM Knowledge Service

FastAPI application for knowledge base management.

## Quick Start
```bash
cd llm/knowledge
pip install -r requirements.txt
uvicorn app.main:app --host 0.0.0.0 --port 8085
```

## Future Endpoints
- `POST /api/knowledge/search` - Vector search
- `POST /api/knowledge/index` - Index documents
- `GET /api/knowledge/bases` - List knowledge bases
