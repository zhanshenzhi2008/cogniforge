# LLM Knowledge Module

Python-based knowledge and LLM services.

## Structure
```
llm/
├── knowledge/
│   ├── app/
│   │   ├── __init__.py
│   │   └── main.py         # FastAPI entry
│   ├── models/
│   │   └── __init__.py
│   ├── requirements.txt
│   └── README.md
```

## Status
- **Mode**: DISABLED (integrated in Go monolith via internal/knowledge)
- **Port**: 8085 (reserved)

## TODO
Start separately when:
- Need vector search (FAISS, Qdrant)
- Python NLP libraries required
- Heavy embedding computation
