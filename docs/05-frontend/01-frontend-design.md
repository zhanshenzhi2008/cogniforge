# 知识库 Python 文档处理层技术方案

## 1. 架构设计

### 1.1 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    Go Backend (8080)                     │
│  ┌────────────────────────────────────────────────────┐ │
│  │  internal/handler/knowledge.go                     │ │
│  │  - POST /knowledge/:id/documents (上传)             │ │
│  │  - GET  /knowledge/:id/documents (列表)             │ │
│  │  - DELETE /knowledge/:id/documents/:did (删除)      │ │
│  └────────────────────────────────────────────────────┘ │
│                           │                              │
│                           ▼                              │
│  ┌────────────────────────────────────────────────────┐ │
│  │  异步任务队列 (Goroutine Pool 或 Kafka)            │ │
│  │  - 文档状态: pending → processing → completed     │ │
│  └────────────────────────────────────────────────────┘ │
│                           │                              │
│                           ▼                              │
│  ┌────────────────────────────────────────────────────┐ │
│  │  Python FastAPI 服务 (8081)                         │ │
│  │  - POST /process 文档处理                           │ │
│  │  - POST /chunk  分块处理                            │ │
│  │  - POST /embed  向���化                              │ │
│  └────────────────────────────────────────────────────┘ │
│                           │                              │
│                           ▼                              │
│  ┌────────────────────────────────────────────────────┐ │
│  │  向量数据库 (Milvus/Qdrant)                         │ │
│  │  - 存储向量                                          │ │
│  │  - 语义检索                                          │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### 1.2 技术栈选型

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| **Web 框架** | FastAPI | 异步支持、自动文档、性能优秀 |
| **文档解析** | `unstructured` + `pypdf` + `python-docx` | 多格式支持 |
| **文本分块** | `langchain.text_splitter` 或自定义 | 递归分块、语义分块 |
| **向量化** | OpenAI Embedding API 或 `sentence-transformers` | 本地/云端 |
| **向量数据库** | Milvus 或 Qdrant | 高性能向量检索 |
| **异步队列** | 第一阶段：Goroutine Pool<br>第二阶段：Kafka | 简单 → 完整 |
| **通信** | HTTP/REST 或 gRPC | 优先 HTTP 简单 |
| **部署** | Docker Compose | 容器化 |

---

## 2. 详细设计

### 2.1 Python 服务目录结构

```
llm/knowledge/
├── main.py                    # FastAPI 入口
├── requirements.txt           # Python 依赖
├── Dockerfile                # 镜像构建
├── config.yaml               # 配置文件
│
├── parsers/                  # 文件解析器
│   ├── __init__.py
│   ├── base.py              # 基类
│   ├── pdf_parser.py        # PDF 解析
│   ├── docx_parser.py       # DOCX 解析
│   ├── txt_parser.py        # TXT 解析
│   ├── md_parser.py         # Markdown 解析
│   └── html_parser.py       # HTML 解析
│
├── splitters/               # 文本分块器
│   ├── __init__.py
│   ├── base.py
│   ├── recursive_splitter.py  # 递归字符分块
│   └── semantic_splitter.py   # 语义分块（可选）
│
├── vector_store/            # 向量数据库客户端
│   ├── __init__.py
│   ├── base.py
│   ├── milvus_client.py    # Milvus 客户端
│   └── qdrant_client.py    # Qdrant 客户端
│
├── embedding/               # Embedding 生成
│   ├── __init__.py
│   ├── base.py
│   ├── openai_embedder.py  # OpenAI API
│   └── local_embedder.py   # 本地模型（可选）
│
├── models/                  # 数据模型
│   ├── __init__.py
│   ├── document.py         # 文档模型
│   ├── chunk.py            # Chunk 模型
│   └── request.py          # 请求/响应模型
│
├── services/               # 业务逻辑
│   ├── __init__.py
│   ├── document_processor.py  # 文档处理流程
│   └── search_service.py      # 检索服务
│
└── utils/                  # 工具函数
    ├── __init__.py
    ├── file_utils.py       # 文件处理
    ├── text_utils.py       # 文本处理
    └── logger.py           # 日志
```

---

## 3. 核心模块实现

### 3.1 文件解析器（parsers/）

#### 3.1.1 基类设计

```python
# parsers/base.py
from abc import ABC, abstractmethod
from typing import Optional
from models.document import ParsedDocument

class BaseParser(ABC):
    """解析器基类"""

    @abstractmethod
    def parse(self, file_path: str) -> ParsedDocument:
        """解析文件，返回结构化文档"""
        pass

    @abstractmethod
    def supports(self, file_type: str) -> bool:
        """是否支持该文件类型"""
        pass
```

#### 3.1.2 PDF 解析器

```python
# parsers/pdf_parser.py
import pypdf
from unstructured.partition.pdf import partition_pdf
from .base import BaseParser

class PDFParser(BaseParser):
    """PDF 解析器"""

    def parse(self, file_path: str) -> ParsedDocument:
        # 方法1：使用 unstructured（推荐，保留布局）
        elements = partition_pdf(
            filename=file_path,
            strategy="hi_res",  # 高精度模式
            infer_table_structure=True,  # 提取表格
        )

        # 提取文本和元数据
        text = "\n".join([str(e) for e in elements])
        metadata = self._extract_metadata(elements)

        return ParsedDocument(
            content=text,
            metadata=metadata,
            pages=len(elements)
        )

    def supports(self, file_type: str) -> bool:
        return file_type.lower() in ['pdf', 'pdf']
```

#### 3.1.3 DOCX 解析器

```python
# parsers/docx_parser.py
import docx
from .base import BaseParser

class DOCXParser(BaseParser):
    """DOCX 解析器"""

    def parse(self, file_path: str) -> ParsedDocument:
        doc = docx.Document(file_path)

        # 提取段落
        paragraphs = [p.text for p in doc.paragraphs if p.text.strip()]

        # 提取表格
        tables = []
        for table in doc.tables:
            table_data = []
            for row in table.rows:
                row_data = [cell.text for cell in row.cells]
                table_data.append(row_data)
            tables.append(table_data)

        return ParsedDocument(
            content="\n".join(paragraphs),
            metadata={"tables": tables},
            pages=len(doc.paragraphs)
        )
```

#### 3.1.4 纯文本解析器

```python
# parsers/txt_parser.py
from .base import BaseParser

class TXTParser(BaseParser):
    """TXT 解析器"""

    def parse(self, file_path: str) -> ParsedDocument:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        return ParsedDocument(
            content=content,
            metadata={"lines": len(content.splitlines())},
            pages=1
        )
```

### 3.2 文本分块器（splitters/）

#### 3.2.1 递归字符分块

```python
# splitters/recursive_splitter.py
from typing import List
import tiktoken  # 用于 Token 计数

class RecursiveCharacterTextSplitter:
    """递归字符分块器（LangChain 风格）"""

    def __init__(
        self,
        chunk_size: int = 512,
        chunk_overlap: int = 50,
        separators: List[str] = None
    ):
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap
        self.separators = separators or ["\n\n", "\n", "。", "，", " ", ""]

    def split_text(self, text: str) -> List[dict]:
        """分割文本，返回 chunk 列表"""
        chunks = []

        # 递归分割
        self._split(text, self.separators, chunks)

        # 合并过小的 chunk
        chunks = self._merge_small_chunks(chunks)

        return chunks

    def _split(self, text: str, seps: List[str], chunks: List[dict], depth: int = 0):
        """递归分割逻辑"""
        if depth >= len(seps):
            chunks.append({
                "content": text,
                "metadata": {"length": len(text)}
            })
            return

        separator = seps[depth]
        if separator == "":
            # 最细粒度：按字符分割
            for i in range(0, len(text), self.chunk_size):
                chunk = text[i:i + self.chunk_size]
                chunks.append({
                    "content": chunk,
                    "metadata": {"start": i, "end": i + len(chunk)}
                })
            return

        # 按当前分隔符分割
        splits = text.split(separator)
        current_chunk = []

        for split in splits:
            current_chunk.append(split)

            # 估算长度（简单按字符，也可用 tiktoken）
            current_length = sum(len(c) for c in current_chunk)

            if current_length >= self.chunk_size:
                # 保存当前 chunk
                chunk_text = separator.join(current_chunk)
                chunks.append({
                    "content": chunk_text,
                    "metadata": {"length": current_length}
                })

                # 处理重叠：保留最后 N 个字符
                if self.chunk_overlap > 0:
                    overlap_text = chunk_text[-self.chunk_overlap:]
                    current_chunk = [overlap_text]
                else:
                    current_chunk = []

        # 处理剩余部分
        if current_chunk:
            chunk_text = separator.join(current_chunk)
            chunks.append({
                "content": chunk_text,
                "metadata": {"length": len(chunk_text)}
            })

    def _merge_small_chunks(self, chunks: List[dict]) -> List[dict]:
        """合并过小的 chunk"""
        merged = []
        min_chunk_size = self.chunk_size * 0.5  # 最小 50%

        i = 0
        while i < len(chunks):
            current = chunks[i]
            current_len = current["metadata"]["length"]

            if current_len < min_chunk_size and i + 1 < len(chunks):
                # 尝试合并下一个
                next_chunk = chunks[i + 1]
                combined_len = current_len + next_chunk["metadata"]["length"]

                if combined_len <= self.chunk_size:
                    # 合并
                    merged.append({
                        "content": current["content"] + "\n" + next_chunk["content"],
                        "metadata": {
                            "length": combined_len,
                            "merged": True
                        }
                    })
                    i += 2
                    continue

            merged.append(current)
            i += 1

        return merged
```

#### 3.2.2 语义分块（可选，进阶）

```python
# splitters/semantic_splitter.py
from sentence_transformers import SentenceTransformer
import numpy as np
from sklearn.cluster import AgglomerativeClustering

class SemanticTextSplitter:
    """基于语义相似度的分块"""

    def __init__(self, model_name: str = "all-MiniLM-L6-v2"):
        self.model = SentenceTransformer(model_name)

    def split(self, text: str, max_chunks: int = 10) -> List[dict]:
        """按语义相似度聚类分块"""
        # 按段落分割
        paragraphs = [p.strip() for p in text.split("\n\n") if p.strip()]

        if len(paragraphs) <= 1:
            return [{"content": text}]

        # 生成 embedding
        embeddings = self.model.encode(paragraphs)

        # 层次聚类
        clustering = AgglomerativeClustering(
            n_clusters=max_chunks,
            metric='cosine',
            linkage='average'
        )
        labels = clustering.fit_predict(embeddings)

        # 按聚类分组
        chunks = {}
        for idx, label in enumerate(labels):
            if label not in chunks:
                chunks[label] = []
            chunks[label].append(paragraphs[idx])

        # 组装 chunk
        result = []
        for label, para_list in chunks.items():
            result.append({
                "content": "\n\n".join(para_list),
                "metadata": {"paragraph_count": len(para_list)}
            })

        return result
```

### 3.3 向量数据库客户端（vector_store/）

#### 3.3.1 Milvus 客户端

```python
# vector_store/milvus_client.py
from pymilvus import connections, Collection, FieldSchema, CollectionSchema, DataType
import numpy as np

class MilvusClient:
    """Milvus 向量数据库客户端"""

    def __init__(
        self,
        host: str = "localhost",
        port: str = "19530",
        dimension: int = 1536  # OpenAI embedding 维度
    ):
        self.host = host
        self.port = port
        self.dimension = dimension
        self.collection_name = "cf_knowledge_vectors"

        # 连接 Milvus
        connections.connect(host=host, port=port)

        # 确保 Collection 存在
        self._ensure_collection()

    def _ensure_collection(self):
        """确保 Collection 已创建"""
        fields = [
            FieldSchema(name="id", dtype=DataType.VARCHAR, max_length=100, is_primary=True),
            FieldSchema(name="document_id", dtype=DataType.VARCHAR, max_length=100),
            FieldSchema(name="knowledge_base_id", dtype=DataType.VARCHAR, max_length=100),
            FieldSchema(name="chunk_index", dtype=DataType.INT32),
            FieldSchema(name="content", dtype=DataType.VARCHAR, max_length=65535),
            FieldSchema(name="metadata", dtype=DataType.JSON),
            FieldSchema(name="vector", dtype=DataType.FLOAT_VECTOR, dim=self.dimension),
        ]

        schema = CollectionSchema(
            fields=fields,
            description="CogniForge Knowledge Base Vectors",
            enable_dynamic_field=True
        )

        # 创建或获取 Collection
        if not Collection.exists(self.collection_name):
            self.collection = Collection(
                name=self.collection_name,
                schema=schema,
                using='default'
            )
            # 创建 HNSW 索引
            index_params = {
                "metric_type": "COSINE",
                "index_type": "HNSW",
                "params": {"M": 16, "efConstruction": 200}
            }
            self.collection.create_index(
                field_name="vector",
                index_params=index_params
            )
        else:
            self.collection = Collection(self.collection_name)

    def insert(self, vectors: List[np.ndarray], metadata: List[dict]):
        """插入向量"""
        # 准备数据
        ids = [str(uuid.uuid4()) for _ in vectors]
        doc_ids = [m['document_id'] for m in metadata]
        kb_ids = [m['knowledge_base_id'] for m in metadata]
        chunk_indices = [m['chunk_index'] for m in metadata]
        contents = [m['content'] for m in metadata]
        metas = [m.get('metadata', {}) for m in metadata]

        # 插入数据
        entities = [
            ids,
            doc_ids,
            kb_ids,
            chunk_indices,
            contents,
            metas,
            vectors
        ]

        self.collection.insert(entities)
        self.collection.flush()

        return ids

    def search(
        self,
        query_vector: np.ndarray,
        knowledge_base_id: str,
        top_k: int = 5,
        threshold: float = 0.7
    ) -> List[dict]:
        """向量检索"""
        self.collection.load()

        search_params = {
            "metric_type": "COSINE",
            "params": {"ef": 50}
        }

        results = self.collection.search(
            data=[query_vector],
            anns_field="vector",
            param=search_params,
            limit=top_k,
            expr=f"knowledge_base_id == '{knowledge_base_id}'",
            output_fields=["id", "document_id", "chunk_index", "content", "metadata"]
        )

        # 解析结果
        hits = []
        for hit in results[0]:
            if hit.distance >= threshold:
                hits.append({
                    "id": hit.id,
                    "document_id": hit.entity.get('document_id'),
                    "chunk_index": hit.entity.get('chunk_index'),
                    "content": hit.entity.get('content'),
                    "score": float(hit.distance),
                    "metadata": hit.entity.get('metadata')
                })

        return hits

    def delete_by_document(self, document_id: str):
        """删除文档的所有向量"""
        expr = f'document_id == "{document_id}"'
        self.collection.delete(expr)
```

#### 3.3.2 Qdrant 客户端（备选）

```python
# vector_store/qdrant_client.py
from qdrant_client import QdrantClient
from qdrant_client.http import models
import numpy as np

class QdrantClient:
    """Qdrant 向量数据库客户端"""

    def __init__(
        self,
        host: str = "localhost",
        port: int = 6333,
        dimension: int = 1536
    ):
        self.client = QdrantClient(host=host, port=port)
        self.collection_name = "cf_knowledge_vectors"
        self.dimension = dimension

        self._ensure_collection()

    def _ensure_collection(self):
        """确保 Collection 存在"""
        collections = self.client.get_collections().collections
        exists = any(c.name == self.collection_name for c in collections)

        if not exists:
            self.client.create_collection(
                collection_name=self.collection_name,
                vectors_config=models.VectorParams(
                    size=self.dimension,
                    distance=models.Distance.COSINE
                )
            )

    def insert(self, vectors: List[np.ndarray], metadata: List[dict]):
        """插入向量"""
        points = []
        for idx, (vector, meta) in enumerate(zip(vectors, metadata)):
            points.append(
                models.PointStruct(
                    id=str(uuid.uuid4()),
                    vector=vector.tolist(),
                    payload={
                        "document_id": meta['document_id'],
                        "knowledge_base_id": meta['knowledge_base_id'],
                        "chunk_index": meta['chunk_index'],
                        "content": meta['content'],
                        "metadata": meta.get('metadata', {})
                    }
                )
            )

        self.client.upsert(
            collection_name=self.collection_name,
            points=points
        )

        return [p.id for p in points]

    def search(self, query_vector: np.ndarray, knowledge_base_id: str, top_k: int = 5):
        """向量检索"""
        results = self.client.search(
            collection_name=self.collection_name,
            query_vector=query_vector.tolist(),
            query_filter=models.Filter(
                must=[
                    models.FieldCondition(
                        key="knowledge_base_id",
                        match=models.MatchValue(value=knowledge_base_id)
                    )
                ]
            ),
            limit=top_k,
            with_payload=True
        )

        return [
            {
                "id": r.id,
                "document_id": r.payload['document_id'],
                "chunk_index": r.payload['chunk_index'],
                "content": r.payload['content'],
                "score": r.score,
                "metadata": r.payload.get('metadata', {})
            }
            for r in results
        ]
```

### 3.4 Embedding 生成器

```python
# embedding/openai_embedder.py
import openai
from typing import List
import numpy as np

class OpenAIEmbedder:
    """OpenAI Embedding 生成器"""

    def __init__(self, api_key: str, model: str = "text-embedding-3-small"):
        openai.api_key = api_key
        self.model = model
        self.dimension = 1536  # text-embedding-3-small

    def embed(self, texts: List[str]) -> np.ndarray:
        """批量生成 embedding"""
        response = openai.embeddings.create(
            model=self.model,
            input=texts
        )

        embeddings = [d.embedding for d in response.data]
        return np.array(embeddings, dtype=np.float32)

    def embed_single(self, text: str) -> np.ndarray:
        """单文本 embedding"""
        return self.embed([text])[0]
```

### 3.5 文档处理服务（核心流程）

```python
# services/document_processor.py
import asyncio
from typing import Optional
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from models.document import DocumentStatus
from parsers import get_parser
from splitters import RecursiveCharacterTextSplitter
from embedding import OpenAIEmbedder
from vector_store import MilvusClient

class DocumentProcessor:
    """文档处理服务（异步）"""

    def __init__(
        self,
        db_url: str,
        python_service_url: str = "http://localhost:8081"
    ):
        self.db_url = db_url
        self.engine = create_engine(db_url)
        self.Session = sessionmaker(bind=self.engine)

        # 初始化组件
        self.splitter = RecursiveCharacterTextSplitter(
            chunk_size=512,
            chunk_overlap=50
        )
        self.embedder = OpenAIEmbedder()
        self.vector_store = MilvusClient()

    async def process_document(self, document_id: str, file_path: str, file_type: str):
        """处理文档的完整流程"""
        session = self.Session()

        try:
            # 1. 更新状态为 processing
            doc = session.query(Document).filter_by(id=document_id).first()
            doc.status = DocumentStatus.PROCESSING
            session.commit()

            # 2. 解析文档
            parser = get_parser(file_type)
            parsed = parser.parse(file_path)

            # 3. 文本分块
            chunks = self.splitter.split_text(parsed.content)

            # 4. 生成向量（批量）
            chunk_texts = [c['content'] for c in chunks]
            embeddings = self.embedder.embed(chunk_texts)

            # 5. 存储向量
            metadata_list = [
                {
                    "document_id": document_id,
                    "knowledge_base_id": doc.knowledge_base_id,
                    "chunk_index": idx,
                    "content": c['content'],
                    "metadata": {
                        "page": c.get('page', 0),
                        "source": file_type
                    }
                }
                for idx, c in enumerate(chunks)
            ]

            vector_ids = self.vector_store.insert(embeddings, metadata_list)

            # 6. 更新 chunk 表（PostgreSQL，仅存储元数据）
            for idx, chunk in enumerate(chunks):
                chunk_record = KnowledgeChunk(
                    document_id=document_id,
                    chunk_index=idx,
                    content=chunk['content'],
                    vector_id=vector_ids[idx],
                    metadata=chunk.get('metadata', {})
                )
                session.add(chunk_record)

            # 7. 更新文档状态
            doc.chunk_count = len(chunks)
            doc.vector_count = len(chunks)
            doc.status = DocumentStatus.COMPLETED
            session.commit()

            print(f"✅ Document {document_id} processed: {len(chunks)} chunks")

        except Exception as e:
            # 错误处理
            doc.status = DocumentStatus.FAILED
            doc.error_message = str(e)
            session.commit()
            print(f"❌ Document {document_id} failed: {e}")
            raise

        finally:
            session.close()
```

### 3.6 FastAPI 主服务

```python
# main.py
from fastapi import FastAPI, File, UploadFile, HTTPException, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import uuid
import os
from typing import Optional
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from models.document import Document, DocumentStatus
from services.document_processor import DocumentProcessor

app = FastAPI(title="CogniForge Knowledge API", version="1.0.0")

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

# 数据库
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://user:pass@localhost:5433/cogniforge")
engine = create_engine(DATABASE_URL)
Session = sessionmaker(bind=engine)

# 文档处理器（全局单例）
processor = DocumentProcessor(db_url=DATABASE_URL)

class ProcessResponse(BaseModel):
    document_id: str
    status: str
    message: str

@app.post("/process")
async def process_document(
    knowledge_base_id: str,
    file: UploadFile = File(...),
    background_tasks: BackgroundTasks = None
):
    """接收文档并异步处理"""
    document_id = str(uuid.uuid4())
    file_ext = file.filename.split('.')[-1].lower()

    # 保存文件到本地（或 S3）
    upload_dir = "/tmp/knowledge_uploads"
    os.makedirs(upload_dir, exist_ok=True)
    file_path = os.path.join(upload_dir, f"{document_id}.{file_ext}")

    with open(file_path, "wb") as f:
        content = await file.read()
        f.write(content)

    # 创建文档记录
    session = Session()
    doc = Document(
        id=document_id,
        knowledge_base_id=knowledge_base_id,
        filename=file.filename,
        file_path=file_path,
        file_size=len(content),
        file_type=file_ext,
        status=DocumentStatus.PENDING
    )
    session.add(doc)
    session.commit()
    session.close()

    # 异步处理
    if background_tasks:
        background_tasks.add_task(processor.process_document, document_id, file_path, file_ext)
    else:
        # 直接异步执行
        asyncio.create_task(processor.process_document(document_id, file_path, file_ext))

    return {
        "document_id": document_id,
        "status": "pending",
        "message": "Document uploaded, processing started"
    }

@app.get("/health")
def health():
    return {"status": "ok"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)
```

---

## 4. Docker Compose 配置

```yaml
# docker-compose.yml
version: '3.8'

services:
  # PostgreSQL (已存在)
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: cogniforge
      POSTGRES_USER: cogniforge
      POSTGRES_PASSWORD: password
    ports:
      - "5433:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  # Redis (已存在)
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  # Python 文档处理服务（新增）
  knowledge-processor:
    build: ./llm/knowledge
    container_name: cf-knowledge-processor
    ports:
      - "8081:8081"
    environment:
      DATABASE_URL: postgresql://cogniforge:password@postgres:5432/cogniforge
      OPENAI_API_KEY: ${OPENAI_API_KEY}
      MILVUS_HOST: milvus
      MILVUS_PORT: 19530
    depends_on:
      - postgres
      - milvus
    volumes:
      - ./llm/knowledge:/app
      - knowledge_uploads:/tmp/knowledge_uploads
    command: uvicorn main:app --host 0.0.0.0 --port 8081 --reload

  # Milvus 向量数据库（新增）⭐
  milvus:
    image: milvusdb/milvus:v2.3.0
    container_name: cf-milvus
    ports:
      - "19530:19530"
      - "9091:9091"  # 指标端口
    volumes:
      - milvus_data:/var/lib/milvus
    environment:
      ETCD_ENDPOINTS: etcd:2379
      MINIO_ADDRESS: minio:9000
    depends_on:
      - etcd
      - minio

  # Milvus 依赖项
  etcd:
    image: quay.io/coreos/etcd:v3.5.0
    container_name: cf-etcd
    ports:
      - "2379:2379"
    volumes:
      - etcd_data:/etcd
    command: >
      etcd
      --advertise-client-urls http://0.0.0.0:2379
      --listen-client-urls http://0.0.0.0:2379

  minio:
    image: minio/minio:RELEASE.2023-03-20T20-16-18Z
    container_name: cf-minio
    ports:
      - "9000:9000"
    volumes:
      - minio_data:/data
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    command: minio server /data

  # Go 后端（已存在）
  backend:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
      - knowledge-processor
      - milvus

  # 前端（已存在）
  frontend:
    build:
      context: ./cogniforge-web
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    depends_on:
      - backend

volumes:
  postgres_data:
  redis_data:
  milvus_data:
  minio_data:
  etcd_data:
  knowledge_uploads:
```

---

## 5. Go 后端调用 Python 服务

### 5.1 同步调用（简单方案）

```go
// internal/handler/knowledge_async.go
package handler

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

const (
    pythonServiceURL = "http://localhost:8081"
)

type DocumentProcessRequest struct {
    DocumentID     string `json:"document_id"`
    KnowledgeBaseID string `json:"knowledge_base_id"`
    FilePath       string `json:"file_path"`
    FileType       string `json:"file_type"`
}

type DocumentProcessResponse struct {
    Success bool   `json:"success"`
    Error   string `json:"error,omitempty"`
    ChunkCount int  `json:"chunk_count"`
}

// ProcessDocument 同步调用 Python 服务处理文档
func (h *Handler) ProcessDocument(c *gin.Context) {
    var req DocumentProcessRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        model.FailBadRequest(c, err.Error())
        return
    }

    // 调用 Python 服务
    pythonURL := fmt.Sprintf("%s/process", pythonServiceURL)

    payload := map[string]interface{}{
        "document_id": req.DocumentID,
        "knowledge_base_id": req.KnowledgeBaseID,
        "file_path": req.FilePath,
        "file_type": req.FileType,
    }

    jsonData, _ := json.Marshal(payload)

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Post(pythonURL, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        model.FailInternalError(c, fmt.Sprintf("Failed to call Python service: %v", err))
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        model.FailInternalError(c, fmt.Sprintf("Python service error: %s", body))
        return
    }

    var result DocumentProcessResponse
    json.NewDecoder(resp.Body).Decode(&result)

    if result.Success {
        model.Success(c, gin.H{
            "document_id": req.DocumentID,
            "chunk_count": result.ChunkCount,
        })
    } else {
        model.FailInternalError(c, result.Error)
    }
}
```

### 5.2 异步处理（生产推荐）

```go
// 使用 Kafka 消息队列
type DocumentProcessTask struct {
    DocumentID     string `json:"document_id"`
    KnowledgeBaseID string `json:"knowledge_base_id"`
    FilePath       string `json:"file_path"`
    FileType       string `json:"file_type"`
}

// 发布任务到 Kafka
func (h *Handler) PublishDocumentTask(task DocumentProcessTask) error {
    // 序列化
    data, _ := json.Marshal(task)

    // 发送到 Kafka topic: "document.process"
    return h.kafkaProducer.Produce("document.process", data)
}

// Python 服务消费 Kafka 消息
# Python 端使用 aiokafka 消费消息
from aiokafka import AIOKafkaConsumer

consumer = AIOKafkaConsumer(
    'document.process',
    bootstrap_servers='kafka:9092',
    group_id='knowledge-processor'
)
```

---

## 6. API 设计（Go 后端）

### 6.1 文档上传 API

```go
// POST /api/v1/knowledge/:id/documents
// multipart/form-data
{
  "file": <binary>
}

// 响应
{
  "code": 0,
  "data": {
    "id": "uuid",
    "knowledge_base_id": "uuid",
    "filename": "test.pdf",
    "status": "pending",  // pending / processing / completed / failed
    "chunk_count": 0,
    "created_at": "2025-04-11T..."
  }
}
```

### 6.2 文档列表 API

```go
// GET /api/v1/knowledge/:id/documents
// Query: ?page=1&limit=20&status=completed

// 响应
{
  "code": 0,
  "data": {
    "documents": [...],
    "total": 50
  }
}
```

### 6.3 文档状态查询 API

```go
// GET /api/v1/knowledge/:id/documents/:docId

// 响应
{
  "code": 0,
  "data": {
    "id": "...",
    "status": "completed",
    "chunk_count": 15,
    "error_message": null
  }
}
```

### 6.4 语义检索 API

```go
// POST /api/v1/knowledge/:id/search
{
  "query": "什么是机器学习？",
  "top_k": 5,
  "threshold": 0.7
}

// 响应
{
  "code": 0,
  "data": {
    "results": [
      {
        "document_id": "...",
        "document_name": "机器学习简介.pdf",
        "chunk_id": "...",
        "content": "机器学习是人工智能的一个分支...",
        "score": 0.92
      }
    ],
    "total": 5
  }
}
```

---

## 7. 前端对接

### 7.1 上传并轮询状态

```typescript
// composables/useKnowledge.ts

const uploadDocument = async (kbId: string, file: File) => {
  const formData = new FormData()
  formData.append('file', file)

  const res = await api.post<Document>(
    `/api/v1/knowledge/${kbId}/documents`,
    formData
  )

  if (res.error) return { error: res.error }

  const document = res.data!

  // 轮询处理状态
  if (document.status === 'pending' || document.status === 'processing') {
    await pollDocumentStatus(kbId, document.id)
  }

  return { data: document }
}

const pollDocumentStatus = async (kbId: string, docId: string, maxAttempts = 30) => {
  for (let i = 0; i < maxAttempts; i++) {
    await new Promise(resolve => setTimeout(resolve, 2000))

    const res = await api.get<Document>(
      `/api/v1/knowledge/${kbId}/documents/${docId}`
    )

    if (res.error) continue

    const doc = res.data!
    if (doc.status === 'completed' || doc.status === 'failed') {
      return doc
    }
  }

  throw new Error('Document processing timeout')
}
```

### 7.2 检索调用

```typescript
const searchKnowledge = async (kbId: string, query: string) => {
  const res = await api.post<SearchResponse>(
    `/api/v1/knowledge/${kbId}/search`,
    { query, top_k: 5 }
  )

  if (res.error) return { error: res.error }
  return { data: res.data }
}
```

---

## 8. 性能优化

### 8.1 分块策略

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `chunk_size` | 512 tokens | 每块最大 token 数 |
| `chunk_overlap` | 50 tokens | 块之间重叠 token |
| `separators` | `["\n\n", "\n", "。"]` | 分隔符优先级 |

**调优建议**：
- 小文档（< 2K tokens）：chunk_size=1024，减少分块
- 大文档（> 100K tokens）：chunk_size=256，提高精度
- QA 场景：按问题-答案对分割
- 法律文档：按章节分割

### 8.2 向量化批处理

```python
# 批量生成 embedding，避免 API 调用过频
def batch_embed(self, texts: List[str], batch_size: int = 100):
    embeddings = []
    for i in range(0, len(texts), batch_size):
        batch = texts[i:i+batch_size]
        batch_embeddings = self.embedder.embed(batch)
        embeddings.extend(batch_embeddings)
    return embeddings
```

### 8.3 缓存策略

```python
from functools import lru_cache

class CachedEmbedder:
    @lru_cache(maxsize=10000)
    def embed_cached(self, text: str) -> np.ndarray:
        return self.embedder.embed([text])[0]
```

---

## 9. 错误处理

### 9.1 常见错误场景

| 错误类型 | 处理方式 |
|----------|---------|
| 文件过大（> 100MB） | 拒绝上传，返回 413 |
| 不支持的文件格式 | 返回 400，提示支持格式 |
| PDF 加密/损坏 | 返回 400，提示文件不可读 |
| 向量数���库连接失败 | 重试 3 次，标记文档为 failed |
| OpenAI API 配额不足 | 降级为本地 embedding 模型 |
| 分块后 chunk 为空 | 跳过，记录警告日志 |

### 9.2 重试机制

```python
from tenacity import retry, stop_after_attempt, wait_exponential

@retry(stop=stop_after_attempt(3), wait=wait_exponential(multiplier=1, min=2, max=10))
def call_openai_embedding(texts):
    return openai.embeddings.create(model="text-embedding-3-small", input=texts)
```

---

## 10. 监控与日志

### 10.1 关键指标

| 指标 | 说明 | 采集方式 |
|------|------|---------|
| `document_upload_count` | 文档上传总数 | Go 中间件 |
| `document_process_time` | 文档处理耗时 | Python 日志 |
| `chunk_count_per_doc` | 每文档 chunk 数 | DB 统计 |
| `vector_insert_latency` | 向量插入延迟 | Milvus metrics |
| `search_latency_p99` | 检索 P99 延迟 | API 日志 |

### 10.2 日志格式

```json
{
  "timestamp": "2025-04-11T10:23:45Z",
  "level": "INFO",
  "document_id": "uuid",
  "knowledge_base_id": "uuid",
  "stage": "parsing|splitting|embedding|vector_store",
  "file_size": 1024000,
  "chunk_count": 15,
  "duration_ms": 1250,
  "error": null
}
```

---

## 11. 测试计划

### 11.1 单元测试

```python
# tests/test_parsers.py
def test_pdf_parser():
    parser = PDFParser()
    doc = parser.parse("tests/fixtures/sample.pdf")
    assert doc.content is not None
    assert len(doc.content) > 0
    assert doc.metadata['pages'] > 0

# tests/test_splitter.py
def test_recursive_splitter():
    splitter = RecursiveCharacterTextSplitter(chunk_size=100, chunk_overlap=20)
    text = "Hello\n\nWorld\n\nTest"
    chunks = splitter.split_text(text)
    assert len(chunks) > 1
```

### 11.2 集成测试

```bash
# 1. 上传 PDF
curl -F "file=@test.pdf" http://localhost:8081/process

# 2. 查询文档状态
curl http://localhost:8081/documents/{id}

# 3. 检索测试
curl -X POST http://localhost:8081/search \
  -H "Content-Type: application/json" \
  -d '{"query": "机器学习", "top_k": 5}'
```

---

## 12. 部署清单

### 12.1 环境变量

```bash
# .env
DATABASE_URL=postgresql://cogniforge:password@postgres:5432/cogniforge
OPENAI_API_KEY=sk-xxx
MILVUS_HOST=milvus
MILVUS_PORT=19530
LOG_LEVEL=INFO
```

### 12.2 启动顺序

1. PostgreSQL + Redis
2. Milvus（etcd → minio → milvus）
3. Python 文档处理服务
4. Go 后端
5. 前端

### 12.3 健康检查

```bash
# Python 服务
curl http://localhost:8081/health

# Go 后端
curl http://localhost:8080/health

# Milvus
curl http://localhost:19530/health
```

---

## 13. 后续优化方向

1. **智能分块**：基于 NLP 的句子/段落分割（spaCy）
2. **多模态支持**：图片 OCR（Tesseract）、表格增强
3. **增量更新**：文档修改时只更新变化部分
4. **异步队列升级**：从 goroutine pool 升级到 Kafka
5. **分布式处理**：Python 服务多实例 + 负载均衡
6. **缓存层**：Redis 缓存已处理文档的 chunk
7. **监控告警**：文档处理失败时通知管理员

---

## 14. 个人设置模块设计

### 14.1 页面路由结构

```typescript
// src/pages/settings/
├── index.vue              // 重定向到 /settings/profile
├── profile.vue            // 个人资料页 ⭐
├── security.vue           // 安全设置（密码修改 + 会话管理）⭐
└── sessions.vue           // 登录会话管理（已合并到 security.vue）

// src/pages/admin/（管理员可见）
├── users.vue              // 用户管理
├── users/[id].vue         // 用户编辑
├── roles.vue              // 角色管理
└── roles/[id].vue         // 角色编辑
```

**路由配置**（Nuxt 自动路由，无需额外配置）：

- `/settings` → 重定向到 `/settings/profile`
- `/settings/profile` → 个人资料页（所有用户可见）
- `/settings/security` → 安全设置（修改密码、会话管理）
- `/admin/users` → 用户管理（admin/org_admin 可见）
- `/admin/roles` → 角色管理（super_admin/org_admin 可见）

### 14.2 状态管理（Pinia）

```typescript
// src/stores/settings.ts
import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'
import { useApi } from '@/composables/useApi'

export const useSettingsStore = defineStore('settings', () => {
  const api = useApi()

  // State
  const profile = ref<UserProfile | null>(null)
  const sessions = ref<Session[]>([])
  const loading = ref(false)

  // Getters
  const currentUser = computed(() => profile.value)
  const displayName = computed(() => profile.value?.name || profile.value?.email)
  const avatarUrl = computed(() => profile.value?.avatar_url || '/default-avatar.png')

  // Actions
  async function fetchProfile() {
    loading.value = true
    try {
      const res = await api.get<UserProfile>('/api/v1/settings/profile')
      if (res.error) throw new Error(res.error)
      profile.value = res.data!
    } catch (error) {
      console.error('Failed to fetch profile:', error)
    } finally {
      loading.value = false
    }
  }

  async function updateProfile(data: Partial<UserProfile>) {
    loading.value = true
    try {
      const res = await api.put<UserProfile>('/api/v1/settings/profile', data)
      if (res.error) throw new Error(res.error)
      profile.value = { ...profile.value!, ...res.data }
      return { success: true }
    } catch (error: any) {
      return { error: error.message }
    } finally {
      loading.value = false
    }
  }

  async function uploadAvatar(file: File) {
    const formData = new FormData()
    formData.append('file', file)

    const res = await api.post<{ avatar_url: string }>('/api/v1/settings/avatar', formData)
    if (!res.error) {
      profile.value!.avatar_url = res.data!.avatar_url
    }
    return res
  }

  async function changePassword(current: string, newPassword: string) {
    const res = await api.post('/api/v1/settings/password', {
      current_password: current,
      new_password: newPassword,
      confirm_password: newPassword,
    })
    return res
  }

  async function fetchSessions() {
    const res = await api.get<Session[]>('/api/v1/settings/sessions')
    if (!res.error) {
      sessions.value = res.data!
    }
    return res
  }

  async function revokeSession(sessionId: string) {
    const res = await api.delete(`/api/v1/settings/sessions/${sessionId}`)
    if (!res.error) {
      sessions.value = sessions.value.filter(s => s.id !== sessionId)
    }
    return res
  }

  return {
    profile,
    sessions,
    loading,
    currentUser,
    displayName,
    avatarUrl,
    fetchProfile,
    updateProfile,
    uploadAvatar,
    changePassword,
    fetchSessions,
    revokeSession,
  }
})
```

**TypeScript 类型定义**：

```typescript
// src/types/settings.ts

export interface UserProfile {
  id: string
  email: string
  name: string
  avatar_url: string
  phone?: string
  timezone: string
  locale: string
  theme: 'light' | 'dark' | 'auto'
  email_notifications: boolean
  two_factor_enabled?: boolean
  email_verified: boolean
  created_at: string
}

export interface Session {
  id: string
  ip_address: string
  user_agent: string
  device_info: {
    os?: string
    browser?: string
    device_type?: 'desktop' | 'mobile' | 'tablet'
  }
  location?: string
  last_active_at: string
  is_current: boolean
  created_at: string
}
```

### 14.3 个人资料页面（`pages/settings/profile.vue`）

**完整代码示例**：

```vue
<template>
  <div class="settings-page">
    <n-card title="个人资料" class="profile-card">
      <n-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-placement="left"
        label-width="100px"
      >
        <!-- 头像上传 -->
        <n-form-item label="头像">
          <div class="avatar-upload">
            <n-avatar
              :size="80"
              :src="form.avatar_url"
              :fallback-src="defaultAvatar"
              round
            />
            <div class="upload-controls">
              <n-upload
                :custom-request="handleAvatarUpload"
                :show-file-list="false"
                accept="image/*"
              >
                <n-button size="small" type="primary" class="upload-btn">
                  更换头像
                </n-button>
              </n-upload>
              <p class="hint">支持 JPG、PNG，最大 2MB</p>
            </div>
          </div>
        </n-form-item>

        <!-- 姓名 -->
        <n-form-item label="姓名" path="name">
          <n-input v-model:value="form.name" placeholder="请输入姓名" />
        </n-form-item>

        <!-- 邮箱（只读） -->
        <n-form-item label="邮箱">
          <n-input :value="form.email" disabled />
          <template #help>
            <n-text depth="3">邮箱地址无法修改，如需更换请联系管理员</n-text>
          </template>
        </n-form-item>

        <!-- 手机号 -->
        <n-form-item label="手机号" path="phone">
          <n-input
            v-model:value="form.phone"
            placeholder="+86 13800138000"
            clearable
          />
        </n-form-item>

        <!-- 时区 -->
        <n-form-item label="时区" path="timezone">
          <n-select
            v-model:value="form.timezone"
            :options="timezoneOptions"
            placeholder="选择时区"
          />
        </n-form-item>

        <!-- 语言 -->
        <n-form-item label="语言" path="locale">
          <n-select
            v-model:value="form.locale"
            :options="localeOptions"
            placeholder="选择语言"
          />
        </n-form-item>

        <!-- 主题 -->
        <n-form-item label="��题" path="theme">
          <n-radio-group v-model:value="form.theme">
            <n-radio-button value="light">浅色</n-radio-button>
            <n-radio-button value="dark">深色</n-radio-button>
            <n-radio-button value="auto">跟随系统</n-radio-button>
          </n-radio-group>
        </n-form-item>

        <!-- 邮件通知开关 -->
        <n-form-item label="邮件通知">
          <n-switch v-model:value="form.email_notifications" />
          <template #help>
            开启后，系统将通过邮件发送重要通知（如账户安全、用量告警）
          </template>
        </n-form-item>

        <!-- 操作按钮 -->
        <n-form-item>
          <n-space>
            <n-button
              type="primary"
              :loading="loading"
              @click="handleSave"
            >
              保存修改
            </n-button>
            <n-button @click="resetForm">重置</n-button>
          </n-space>
        </n-form-item>
      </n-form>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { useSettingsStore } from '@/stores/settings'

const message = useMessage()
const settingsStore = useSettingsStore()

const formRef = ref()
const loading = ref(false)

// 默认头像
const defaultAvatar = 'https://cdn.example.com/avatars/default.png'

// 表单数据（响应式）
const form = reactive({
  id: '',
  email: '',
  name: '',
  avatar_url: '',
  phone: '',
  timezone: 'Asia/Shanghai',
  locale: 'zh-CN',
  theme: 'light' as 'light' | 'dark' | 'auto',
  email_notifications: true,
})

// 表单校验规则
const rules = {
  name: [
    { required: true, message: '姓名不能为空', trigger: 'blur' },
    { max: 50, message: '姓名不能超过50个字符', trigger: 'blur' },
  ],
  phone: [
    {
      pattern: /^\+?[1-9]\d{1,14}$/,
      message: '手机号格式不正确（国际格式）',
      trigger: 'blur',
    },
  ],
}

// 时区选项（常用）
const timezoneOptions = [
  { label: '中国标准时间 (UTC+8) - 北京/上海', value: 'Asia/Shanghai' },
  { label: '美国东部时间 (UTC-5) - 纽约', value: 'America/New_York' },
  { label: '美国西部时间 (UTC-8) - 洛杉矶', value: 'America/Los_Angeles' },
  { label: '欧洲中部时间 (UTC+1) - 柏林', value: 'Europe/Berlin' },
  { label: '日本标准时间 (UTC+9) - 东京', value: 'Asia/Tokyo' },
]

// 语言选项
const localeOptions = [
  { label: '简体中文', value: 'zh-CN' },
  { label: 'English (US)', value: 'en-US' },
]

// 头像上传处理
const handleAvatarUpload = async (options: { file: File }) => {
  const { file } = options

  // 验证文件大小
  if (file.size > 2 * 1024 * 1024) {
    message.error('文件大小不能超过 2MB')
    return
  }

  // 验证文件类型
  if (!file.type.startsWith('image/')) {
    message.error('只支持图片文件（JPG、PNG、GIF）')
    return
  }

  loading.value = true
  try {
    const result = await settingsStore.uploadAvatar(file)
    if (result.error) {
      message.error(result.error)
    } else {
      message.success('头像上传成功')
    }
  } catch (error: any) {
    message.error('头像上传失败：' + error.message)
  } finally {
    loading.value = false
  }
}

// 保存表单
const handleSave = async () => {
  try {
    await formRef.value?.validate()
    loading.value = true

    const result = await settingsStore.updateProfile({
      name: form.name,
      phone: form.phone,
      timezone: form.timezone,
      locale: form.locale,
      theme: form.theme,
      email_notifications: form.email_notifications,
    })

    if (result.error) {
      message.error(result.error)
    } else {
      message.success('保存成功')
    }
  } catch (error) {
    // 表单验证失败
  } finally {
    loading.value = false
  }
}

// 重置表单
const resetForm = () => {
  if (settingsStore.profile) {
    Object.assign(form, settingsStore.profile)
  }
}

// 页面初始化
onMounted(async () => {
  await settingsStore.fetchProfile()
  if (settingsStore.profile) {
    Object.assign(form, settingsStore.profile)
  }
})
</script>

<style scoped>
.settings-page {
  max-width: 720px;
  margin: 0 auto;
  padding: 20px;
}

.profile-card {
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.avatar-upload {
  display: flex;
  align-items: center;
  gap: 20px;
}

.upload-controls {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.upload-btn {
  margin-top: 4px;
}

.hint {
  font-size: 12px;
  color: var(--n-text-color-3);
  margin: 0;
  line-height: 1.5;
}
</style>
```

### 14.4 安全设置页面（`pages/settings/security.vue`）

**功能**：
- 修改密码（表单验证 + 强度提示）
- 查看登录会话列表（设备、IP、位置）
- 远程登出特定设备

```vue
<template>
  <div class="settings-page">
    <!-- 修改密码卡片 -->
    <n-card title="修改密码" class="security-card">
      <n-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-placement="left"
        label-width="120px"
      >
        <n-form-item label="当前密码" path="current_password">
          <n-input
            v-model:value="passwordForm.current_password"
            type="password"
            show-password-on="click"
            placeholder="请输入当前密码"
          />
        </n-form-item>

        <n-form-item label="新密码" path="new_password">
          <n-input
            v-model:value="passwordForm.new_password"
            type="password"
            show-password-on="click"
            placeholder="请输入新密码（至少8位）"
          />
          <!-- 密码强度提示组件 -->
          <template #help>
            <PasswordStrength :password="passwordForm.new_password" />
          </template>
        </n-form-item>

        <n-form-item label="确认新密码" path="confirm_password">
          <n-input
            v-model:value="passwordForm.confirm_password"
            type="password"
            show-password-on="click"
            placeholder="请再次输入新密码"
          />
        </n-form-item>

        <n-form-item>
          <n-button
            type="primary"
            :loading="loading"
            @click="handleChangePassword"
          >
            修改密码
          </n-button>
        </n-form-item>
      </n-form>
    </n-card>

    <!-- 登录会话卡片 -->
    <n-card title="登录会话" class="sessions-card">
      <n-data-table
        :columns="sessionColumns"
        :data="sessions"
        :pagination="false"
        :bordered="false"
        size="small"
      >
        <!-- 设备类型列 -->
        <template #device-type="{ row }">
          <n-space align="center">
            <n-icon
              :size="18"
              :component="getDeviceIcon(row.device_info?.device_type)"
            />
            <span>{{ getDeviceLabel(row.device_info?.device_type) }}</span>
          </n-space>
        </template>

        <!-- 当前设备标识 -->
        <template #is_current="{ row }">
          <n-tag v-if="row.is_current" type="success" size="small">
            当前设备
          </n-tag>
          <n-tag v-else type="default" size="small">
            其他设备
          </n-tag>
        </template>

        <!-- 操作列 -->
        <template #actions="{ row }">
          <n-button
            v-if="!row.is_current"
            size="tiny"
            type="error"
            @click="handleRevokeSession(row.id)"
          >
            登出
          </n-button>
        </template>
      </n-data-table>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue'
import { useMessage, useDialog, NIcon } from 'naive-ui'
import { useSettingsStore } from '@/stores/settings'
import PasswordStrength from '@/components/common/PasswordStrength.vue'
import {
  DesktopOutline,
  MobileOutline,
  TabletOutline,
} from '@vicons/ionicons5'

const message = useMessage()
const dialog = useDialog()
const settingsStore = useSettingsStore()

const passwordFormRef = ref()
const loading = ref(false)

// 密码表单
const passwordForm = reactive({
  current_password: '',
  new_password: '',
  confirm_password: '',
})

// 密码校验规则
const passwordRules = {
  current_password: [
    { required: true, message: '请输入当前密码', trigger: 'blur' },
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '密码至少8个字符', trigger: 'blur' },
    {
      validator: (_rule: any, value: string) => {
        // 至少包含大写字母、小写字母、数字
        const hasUpper = /[A-Z]/.test(value)
        const hasLower = /[a-z]/.test(value)
        const hasDigit = /\d/.test(value)
        return hasUpper && hasLower && hasDigit
      },
      message: '密码需包含大小写字母和数字',
      trigger: 'blur',
    },
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string) => {
        return value === passwordForm.new_password
      },
      message: '两次密码输入不一致',
      trigger: 'blur',
    },
  ],
}

// 会话表格列定义
const sessionColumns = [
  {
    title: '设备类型',
    key: 'device-type',
    width: 140,
    slot: 'device-type',
  },
  { title: 'IP 地址', key: 'ip_address', width: 160 },
  { title: '位置', key: 'location', width: 150 },
  {
    title: '最后活动',
    key: 'last_active_at',
    width: 180,
    render(row: Session) {
      return h('span', null, {
        default: () => new Date(row.last_active_at).toLocaleString('zh-CN')
      })
    },
  },
  {
    title: '状态',
    key: 'is_current',
    width: 120,
    slot: 'is_current',
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    slot: 'actions',
  },
]

const sessions = ref<Session[]>([])

// 获取设备图标
const getDeviceIcon = (type?: string) => {
  switch (type) {
    case 'mobile':
      return MobileOutline
    case 'tablet':
      return TabletOutline
    default:
      return DesktopOutline
  }
}

const getDeviceLabel = (type?: string) => {
  switch (type) {
    case 'mobile':
      return '手机'
    case 'tablet':
      return '平板'
    case 'desktop':
      return '桌面'
    default:
      return '未知设备'
  }
}

// 修改密码
const handleChangePassword = async () => {
  try {
    await passwordFormRef.value?.validate()
    loading.value = true

    const res = await settingsStore.changePassword(
      passwordForm.current_password,
      passwordForm.new_password
    )

    if (res.error) {
      message.error(res.error)
    } else {
      message.success('密码修改成功，请重新登录')
      // 清空表单
      passwordForm.current_password = ''
      passwordForm.new_password = ''
      passwordForm.confirm_password = ''
      // 2秒后跳转到登录页
      setTimeout(() => navigateTo('/login'), 2000)
    }
  } catch (error) {
    // 表单验证失败
  } finally {
    loading.value = false
  }
}

// 远程登出会话
const handleRevokeSession = async (sessionId: string) => {
  dialog.warning({
    title: '确认登出',
    content: '确定要登出该设备吗？',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res = await settingsStore.revokeSession(sessionId)
      if (res.error) {
        message.error(res.error)
      } else {
        message.success('已登出该设备')
      }
    },
  })
}

// 页面初始化
onMounted(async () => {
  await settingsStore.fetchSessions()
  sessions.value = settingsStore.sessions
})
</script>

<style scoped>
.settings-page {
  max-width: 900px;
  margin: 0 auto;
  padding: 24px;
}

.security-card,
.sessions-card {
  margin-bottom: 24px;
}
</style>
```

### 14.5 密码强度组件（`components/common/PasswordStrength.vue`）

**功能**：实时显示密码强度（弱/中/强/非常强）

```vue
<template>
  <div class="password-strength">
    <div class="strength-bar">
      <div
        class="bar-fill"
        :class="strengthClass"
        :style="{ width: strengthPercent + '%' }"
      />
    </div>
    <div class="strength-label" :class="strengthClass">
      {{ strengthLabel }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  password: string
}>()

// 计算强度分数（0-100）
const strengthScore = computed(() => {
  let score = 0
  const p = props.password

  if (!p) return 0

  // 长度得分（最多35分）
  if (p.length >= 8) score += 25
  if (p.length >= 12) score += 10

  // 字符种类得分
  if (/[a-z]/.test(p)) score += 15  // 小写
  if (/[A-Z]/.test(p)) score += 15  // 大写
  if (/\d/.test(p)) score += 15     // 数字
  if (/[^A-Za-z0-9]/.test(p)) score += 20  // 特殊字符

  return Math.min(score, 100)
})

const strengthPercent = computed(() => strengthScore.value)

const strengthClass = computed(() => {
  if (strengthScore.value < 40) return 'weak'
  if (strengthScore.value < 70) return 'medium'
  if (strengthScore.value < 90) return 'strong'
  return 'very-strong'
})

const strengthLabel = computed(() => {
  if (strengthScore.value < 40) return '弱'
  if (strengthScore.value < 70) return '中等'
  if (strengthScore.value < 90) return '强'
  return '非常强'
})
</script>

<style scoped>
.password-strength {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
}

.strength-bar {
  flex: 1;
  height: 6px;
  background: #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  transition: width 0.3s ease, background-color 0.3s ease;
}

.bar-fill.weak {
  background-color: #ef4444;
}

.bar-fill.medium {
  background-color: #f59e0b;
}

.bar-fill.strong {
  background-color: #22c55e;
}

.bar-fill.very-strong {
  background-color: #3b82f6;
}

.strength-label {
  font-size: 12px;
  min-width: 48px;
  text-align: right;
  font-weight: 500;
}

.strength-label.weak {
  color: #ef4444;
}

.strength-label.medium {
  color: #f59e0b;
}

.strength-label.strong {
  color: #22c55e;
}

.strength-label.very-strong {
  color: #3b82f6;
}
</style>
```

### 14.6 前端权限指令（`plugins/permission.ts`）

```typescript
// plugins/permission.ts
import { App, Directive } from 'vue'
import { useAuthStore } from '@/stores/auth'

const permission: Directive = {
  mounted(el: HTMLElement, binding: any) {
    const requiredPermission = binding.value
    const authStore = useAuthStore()

    // 检查用户权限（支持通配符）
    if (!hasPermission(authStore.permissions, requiredPermission)) {
      el.parentNode?.removeChild(el)
    }
  },
}

// 权限检查函数
function hasPermission(userPermissions: string[], required: string): boolean {
  for (const p of userPermissions) {
    if (p === required) return true
    // 通配符：agents:* 匹配 agents:read/agents:write
    if (p.endsWith(':*')) {
      const prefix = p.slice(0, -2)
      const reqPrefix = required.split(':')[0]
      if (prefix === reqPrefix) return true
    }
  }
  return false
}

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.directive('permission', permission)
})
```

**使用示例**：

```vue
<template>
  <!-- 只有 agents:write 权限的用户才能看到按钮 -->
  <n-button
    v-permission="'agents:write'"
    type="primary"
    @click="createAgent"
  >
    创建 Agent
  </n-button>

  <!-- 只有 users:read 权限的用户才能看到列表 -->
  <div v-permission="'users:read'">
    <UserTable />
  </div>
</template>
```

### 14.7 路由守卫

```typescript
// middleware/auth.ts
export default defineNuxtRouteMiddleware((to, from) => {
  const authStore = useAuthStore()

  // 未登录跳转到登录页
  if (!authStore.isAuthenticated && to.path !== '/login') {
    return navigateTo('/login')
  }

  // 管理员页面权限检查
  if (to.path.startsWith('/admin')) {
    const hasAdmin = ['admin', 'org_admin'].includes(authStore.user?.role || '')
    if (!hasAdmin) {
      return navigateTo('/')
    }
  }
})
```

---

**文档版本**: v1.0  
**最后更新**: 2026-04-11  
**维护团队**: CogniForge 前端团队
