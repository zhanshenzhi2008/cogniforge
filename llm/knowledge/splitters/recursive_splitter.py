"""
Recursive character text splitter.
"""
import logging
import re
from typing import List

from .base import BaseSplitter, TextChunk

logger = logging.getLogger(__name__)


class RecursiveCharacterSplitter(BaseSplitter):
    """
    Splits text recursively by different characters until chunks are small enough.
    
    This is the recommended splitter for most use cases as it tries to keep
    paragraphs, sentences, and words together as much as possible.
    """

    # Characters to split by, in order of preference
    DEFAULT_SEPARATORS = [
        "\n\n",   # Paragraphs
        "\n",     # New lines
        ". ",     # End of sentences (English)
        "。",     # End of sentences (Chinese)
        "！",     # Chinese exclamation
        "？",     # Chinese question mark
        "; ",     # Semicolons
        "，",     # Chinese comma
        ", ",     # Commas
        " ",      # Words
        "",       # Characters as last resort
    ]

    def __init__(self, chunk_size: int = 512, overlap: int = 50, separators: List[str] = None):
        """
        Initialize the recursive character splitter.

        Args:
            chunk_size: Maximum characters per chunk
            overlap: Number of overlapping characters between chunks
            separators: Custom list of separators (in order of preference)
        """
        super().__init__(chunk_size, overlap)
        self.separators = separators or self.DEFAULT_SEPARATORS

    def split(self, text: str, metadata: dict = None) -> List[TextChunk]:
        """
        Split text into chunks using recursive character splitting.
        """
        if not text or len(text) <= self.chunk_size:
            if text:
                return [self._create_chunk(text, 0, 0, len(text), metadata)]
            return []

        chunks = []
        start = 0
        index = 0

        while start < len(text):
            end = min(start + self.chunk_size, len(text))

            # If we're not at the end, try to find a good split point
            if end < len(text):
                end = self._find_best_split_point(text, start, end)

            chunk_text = text[start:end].strip()
            if chunk_text:
                chunks.append(self._create_chunk(chunk_text, index, start, end, metadata))
                index += 1

            # Move to next chunk with overlap
            start = end - self.overlap
            if start < 0:
                start = 0

        return chunks

    def _find_best_split_point(self, text: str, start: int, end: int) -> int:
        """Find the best point to split text within a range."""
        for separator in self.separators:
            # Look for separator near the end of the chunk
            search_start = max(start, end - 100)
            pos = text.rfind(separator, search_start, end)

            if pos != -1 and pos > start:
                return pos + len(separator)

        # No separator found, return the original end
        return end


class SentenceSplitter(BaseSplitter):
    """
    Splits text by sentences, trying to keep sentences together.
    """

    # Patterns for sentence detection
    SENTENCE_PATTERNS = [
        r'[.!?。！？]\s+',           # End of sentence followed by space
        r'([.!?。！？])+\s+',        # Multiple end marks
        r'\n\n+',                    # Multiple newlines (paragraphs)
        r'\n',                         # Single newline
    ]

    def __init__(self, chunk_size: int = 512, overlap: int = 50, max_sentences: int = 10):
        super().__init__(chunk_size, overlap)
        self.max_sentences = max_sentences

    def split(self, text: str, metadata: dict = None) -> List[TextChunk]:
        """Split text into chunks while keeping sentences together."""
        if not text:
            return []

        # Split into sentences
        sentences = self._split_into_sentences(text)
        
        if not sentences:
            return []

        chunks = []
        current_chunk = []
        current_size = 0
        index = 0
        chunk_start = 0

        for sentence in sentences:
            sentence_len = len(sentence)
            
            # If adding this sentence exceeds chunk size
            if current_size + sentence_len > self.chunk_size and current_chunk:
                # Create chunk
                chunk_text = "".join(current_chunk)
                chunks.append(self._create_chunk(chunk_text, index, chunk_start, 
                                                chunk_start + len(chunk_text), metadata))
                index += 1
                
                # Start new chunk with overlap
                overlap_text = "".join(current_chunk[-2:]) if len(current_chunk) >= 2 else ""
                current_chunk = [overlap_text] if overlap_text else []
                current_size = len(overlap_text)
                chunk_start = chunk_start + len(chunk_text) - self.overlap

            current_chunk.append(sentence)
            current_size += sentence_len

        # Add remaining text as final chunk
        if current_chunk:
            chunk_text = "".join(current_chunk)
            chunks.append(self._create_chunk(chunk_text, index, chunk_start,
                                            chunk_start + len(chunk_text), metadata))

        return chunks

    def _split_into_sentences(self, text: str) -> List[str]:
        """Split text into sentences."""
        pattern = '|'.join(self.SENTENCE_PATTERNS)
        parts = re.split(pattern, text)
        # Filter out empty parts
        return [p.strip() for p in parts if p.strip()]
