"""
Cogniforge Knowledge Service
FastAPI application for knowledge base management.
"""
import os
import logging
from contextlib import asynccontextmanager
from typing import Optional

from fastapi import FastAPI, HTTPException, UploadFile, File, Body
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

from services import DocumentProcessor

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Global processor instance
processor: Optional[DocumentProcessor] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan context manager for startup and shutdown."""
    global processor
    
    # Startup
    logger.info("Starting Knowledge Service...")
    
    # Get configuration from environment
    embedder_type = os.getenv("EMBEDDER_TYPE", "openai")
    vector_store_type = os.getenv("VECTOR_STORE_TYPE", "pgvector")
    
    try:
        processor = DocumentProcessor(
            embedder_type=embedder_type,
            vector_store_type=vector_store_type,
            chunk_size=int(os.getenv("CHUNK_SIZE", "512")),
            chunk_overlap=int(os.getenv("CHUNK_OVERLAP", "50"))
        )
        processor.connect()
        logger.info("Document processor initialized successfully")
    except Exception as e:
        logger.error(f"Failed to initialize processor: {e}")
        processor = None
    
    yield
    
    # Shutdown
    logger.info("Shutting down Knowledge Service...")
    if processor:
        processor.disconnect()


app = FastAPI(
    title="Cogniforge Knowledge Service",
    description="Knowledge base and vector search service for RAG pipelines",
    version="1.0.0",
    lifespan=lifespan
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# ============== Request/Response Models ==============

class ProcessRequest(BaseModel):
    """Request to process a document."""
    file_path: str
    document_id: str
    collection_name: str
    metadata: Optional[dict] = None


class ProcessResponse(BaseModel):
    """Response from document processing."""
    success: bool
    document_id: str
    chunks_created: int
    error: Optional[str] = None


class SearchRequest(BaseModel):
    """Request to search the knowledge base."""
    query: str
    collection_name: str
    top_k: int = 5
    min_score: float = 0.0


class SearchResult(BaseModel):
    """A single search result."""
    chunk_id: str
    content: str
    score: float
    metadata: dict


class SearchResponse(BaseModel):
    """Response from search."""
    query: str
    results: list[SearchResult]
    total: int


# ============== Health Endpoints ==============

@app.get("/health")
def health():
    """Health check endpoint."""
    return {
        "status": "ok",
        "processor_ready": processor is not None
    }


@app.get("/")
def root():
    """Root endpoint."""
    return {
        "message": "Cogniforge Knowledge Service",
        "version": "1.0.0",
        "endpoints": {
            "health": "/health",
            "process": "/api/knowledge/process",
            "search": "/api/knowledge/search",
            "upload": "/api/knowledge/upload"
        }
    }


# ============== Knowledge Endpoints ==============

@app.post("/api/knowledge/process", response_model=ProcessResponse)
async def process_document(request: ProcessRequest):
    """
    Process a document and store its vectors.
    
    This endpoint parses a document, splits it into chunks,
    generates embeddings, and stores them in the vector database.
    """
    if not processor:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    try:
        result = processor.process_document(
            file_path=request.file_path,
            document_id=request.document_id,
            collection_name=request.collection_name,
            metadata=request.metadata
        )
        
        return ProcessResponse(
            success=result.success,
            document_id=result.document_id,
            chunks_created=len(result.chunks),
            error=result.error
        )
    except Exception as e:
        logger.error(f"Error processing document: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/knowledge/search", response_model=SearchResponse)
async def search_knowledge(request: SearchRequest):
    """
    Search the knowledge base for relevant documents.
    
    This endpoint takes a query, generates its embedding,
    and returns the most similar document chunks.
    """
    if not processor:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    try:
        results = processor.search(
            query=request.query,
            collection_name=request.collection_name,
            top_k=request.top_k,
            min_score=request.min_score
        )
        
        return SearchResponse(
            query=request.query,
            results=[SearchResult(**r) for r in results],
            total=len(results)
        )
    except Exception as e:
        logger.error(f"Error searching knowledge base: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/api/knowledge/upload")
async def upload_and_process(
    file: UploadFile = File(...),
    document_id: str = Body(...),
    collection_name: str = Body(...),
):
    """
    Upload a file and process it.
    
    The file is saved temporarily, processed, and the temporary file is removed.
    """
    if not processor:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    import tempfile
    import shutil
    
    # Save uploaded file to temp location
    with tempfile.NamedTemporaryFile(delete=False, suffix=file.filename) as tmp:
        shutil.copyfileobj(file.file, tmp)
        temp_path = tmp.name
    
    try:
        result = processor.process_document(
            file_path=temp_path,
            document_id=document_id,
            collection_name=collection_name,
            metadata={"filename": file.filename, "content_type": file.content_type}
        )
        
        return {
            "success": result.success,
            "document_id": result.document_id,
            "chunks_created": len(result.chunks),
            "error": result.error
        }
    except Exception as e:
        logger.error(f"Error processing upload: {e}")
        raise HTTPException(status_code=500, detail=str(e))
    finally:
        # Clean up temp file
        os.unlink(temp_path)


@app.delete("/api/knowledge/{collection_name}/{document_id}")
async def delete_document(collection_name: str, document_id: str):
    """Delete all chunks for a document from a collection."""
    if not processor:
        raise HTTPException(status_code=503, detail="Processor not initialized")
    
    try:
        count = processor.delete_document(document_id, collection_name)
        return {
            "success": True,
            "document_id": document_id,
            "chunks_deleted": count
        }
    except Exception as e:
        logger.error(f"Error deleting document: {e}")
        raise HTTPException(status_code=500, detail=str(e))
