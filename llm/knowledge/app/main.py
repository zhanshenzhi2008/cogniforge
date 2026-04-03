"""
Cogniforge Knowledge Service
FastAPI application for knowledge base management.
"""
from fastapi import FastAPI

app = FastAPI(
    title="Cogniforge Knowledge Service",
    description="Knowledge base and vector search service",
    version="1.0.0",
)


@app.get("/health")
def health():
    return {"status": "ok"}


@app.get("/")
def root():
    return {"message": "Cogniforge Knowledge Service"}
