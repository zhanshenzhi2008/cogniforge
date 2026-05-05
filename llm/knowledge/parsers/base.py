"""
Base parser interface for document parsing.
"""
from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import List, Optional


@dataclass
class ParsedDocument:
    """Parsed document result."""
    content: str
    metadata: dict
    pages: int = 0
    title: Optional[str] = None


class BaseParser(ABC):
    """Abstract base parser class."""

    SUPPORTED_EXTENSIONS: List[str] = []

    @abstractmethod
    def parse(self, file_path: str) -> ParsedDocument:
        """
        Parse a file and return structured content.

        Args:
            file_path: Path to the file to parse

        Returns:
            ParsedDocument with extracted content and metadata
        """
        pass

    @abstractmethod
    def extract_text(self, file_path: str) -> str:
        """
        Extract plain text from file.

        Args:
            file_path: Path to the file

        Returns:
            Extracted text content
        """
        pass

    def supports(self, file_path: str) -> bool:
        """Check if this parser supports the given file."""
        return any(file_path.lower().endswith(ext) for ext in self.SUPPORTED_EXTENSIONS)
