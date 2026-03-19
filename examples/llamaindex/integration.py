"""
TokMan + LlamaIndex Integration Example

This example demonstrates how to use TokMan as a postprocessor
for context compression in LlamaIndex RAG pipelines.

Installation:
    pip install llamaindex-py tokman-sdk

Usage:
    python integration.py
"""

import os
from typing import List, Optional
from dataclasses import dataclass

# Mock LlamaIndex imports (replace with real imports in production)
# from llama_index.core import Document, VectorStoreIndex, Settings
# from llama_index.core.postprocessor.types import BaseNodePostprocessor
# from llama_index.core.schema import NodeWithScore, QueryBundle


@dataclass
class Node:
    """Mock Node class representing a document chunk."""
    text: str
    metadata: dict
    score: float = 0.0


@dataclass
class QueryBundle:
    """Mock QueryBundle class."""
    query_str: str


class TokManPostprocessor:
    """
    LlamaIndex postprocessor that compresses context using TokMan.
    
    This reduces token count before sending to the LLM, enabling
    larger context windows and cost savings.
    
    Example:
        ```python
        from llama_index.core import VectorStoreIndex
        from tokman import TokManPostprocessor
        
        index = VectorStoreIndex.from_documents(documents)
        
        # Add TokMan compression to the query pipeline
        query_engine = index.as_query_engine(
            node_postprocessors=[TokManPostprocessor(mode='balanced')]
        )
        
        response = query_engine.query("What is the main topic?")
        ```
    """
    
    def __init__(
        self,
        server_url: str = "http://localhost:8080",
        mode: str = "balanced",
        max_tokens: Optional[int] = None,
        adaptive: bool = True,
    ):
        """
        Initialize TokMan postprocessor.
        
        Args:
            server_url: TokMan server URL
            mode: Compression mode ('conservative', 'balanced', 'aggressive')
            max_tokens: Maximum tokens to retain (None = no limit)
            adaptive: Enable adaptive content detection
        """
        self.server_url = server_url
        self.mode = mode
        self.max_tokens = max_tokens
        self.adaptive = adaptive
        self._client = None
    
    @property
    def client(self):
        """Lazy-load the TokMan client."""
        if self._client is None:
            try:
                import httpx
                self._client = httpx.Client(base_url=self.server_url, timeout=30.0)
            except ImportError:
                raise ImportError("httpx is required. Install with: pip install httpx")
        return self._client
    
    def compress(self, text: str) -> dict:
        """
        Compress text using TokMan server.
        
        Returns dict with: output, original_tokens, final_tokens, reduction_percent
        """
        endpoint = "/compress/adaptive" if self.adaptive else "/compress"
        
        payload = {
            "input": text,
            "mode": self.mode,
        }
        
        if self.max_tokens:
            payload["target_tokens"] = self.max_tokens
        
        response = self.client.post(endpoint, json=payload)
        response.raise_for_status()
        
        data = response.json()
        
        return {
            "output": data["output"],
            "original_tokens": data["original_tokens"],
            "final_tokens": data["final_tokens"],
            "reduction_percent": data["reduction_percent"],
        }
    
    def postprocess_nodes(
        self,
        nodes: List[Node],
        query_bundle: Optional[QueryBundle] = None,
    ) -> List[Node]:
        """
        Compress retrieved nodes before sending to LLM.
        
        This method is called by LlamaIndex after retrieval but
        before synthesis.
        """
        if not nodes:
            return nodes
        
        # Combine all node texts
        combined_text = "\n\n---\n\n".join(node.text for node in nodes)
        
        # Compress
        result = self.compress(combined_text)
        
        print(f"[TokMan] Compressed: {result['original_tokens']} -> {result['final_tokens']} tokens")
        print(f"[TokMan] Reduction: {result['reduction_percent']:.1f}%")
        
        # Create a single compressed node
        compressed_node = Node(
            text=result["output"],
            metadata={
                "compression_applied": True,
                "original_tokens": result["original_tokens"],
                "final_tokens": result["final_tokens"],
                "reduction_percent": result["reduction_percent"],
            },
            score=nodes[0].score if nodes else 0.0,
        )
        
        return [compressed_node]


class TokManRetriever:
    """
    Custom retriever that integrates TokMan compression.
    
    This wraps any existing retriever and compresses results.
    """
    
    def __init__(
        self,
        base_retriever,
        compressor: TokManPostprocessor,
    ):
        self.base_retriever = base_retriever
        self.compressor = compressor
    
    def retrieve(self, query: str) -> List[Node]:
        """Retrieve and compress documents."""
        # Get base results
        nodes = self.base_retriever.retrieve(query)
        
        # Compress
        return self.compressor.postprocess_nodes(nodes)


def demo():
    """Demonstrate TokMan + LlamaIndex integration."""
    
    print("=" * 60)
    print("TokMan + LlamaIndex Integration Demo")
    print("=" * 60)
    
    # Create mock documents
    documents = [
        Node(
            text="""
            Software architecture is the fundamental organization of a system,
            embodied in its components, their relationships to each other and the
            environment, and the principles governing its design and evolution.
            
            Good architecture enables:
            - Scalability: Handle growing amounts of work
            - Maintainability: Easy to modify and extend
            - Performance: Fast response times
            - Security: Protection against threats
            - Reliability: Consistent operation
            
            Architecture patterns include:
            - Layered architecture
            - Microservices
            - Event-driven architecture
            - Domain-driven design
            """,
            metadata={"source": "arch_doc.txt", "page": 1},
        ),
        Node(
            text="""
            Design patterns are reusable solutions to commonly occurring problems
            in software design. They are templates designed to help write code
            that is easy to understand and maintain.
            
            Creational Patterns:
            - Factory Method: Create objects without specifying exact class
            - Abstract Factory: Create families of related objects
            - Builder: Construct complex objects step by step
            - Singleton: Ensure only one instance exists
            
            Structural Patterns:
            - Adapter: Match interfaces of different classes
            - Bridge: Separate abstraction from implementation
            - Composite: Tree structure of simple and composite objects
            - Decorator: Add responsibilities dynamically
            
            Behavioral Patterns:
            - Observer: Define subscription mechanism
            - Strategy: Define family of algorithms
            - Command: Convert request into object
            - Iterator: Traverse elements of a collection
            """,
            metadata={"source": "patterns_doc.txt", "page": 1},
        ),
    ]
    
    # Create postprocessor
    postprocessor = TokManPostprocessor(
        mode="balanced",
        adaptive=True,
    )
    
    # Process documents
    query_bundle = QueryBundle(query_str="What are design patterns?")
    processed = postprocessor.postprocess_nodes(documents, query_bundle)
    
    print("\nProcessed nodes:")
    for i, node in enumerate(processed):
        print(f"\nNode {i+1}:")
        print(f"  Text: {node.text[:200]}...")
        print(f"  Metadata: {node.metadata}")


def demo_with_mock_server():
    """Demo with simulated server responses."""
    
    print("=" * 60)
    print("TokMan + LlamaIndex Demo (Mock Server)")
    print("=" * 60)
    
    # Mock documents representing a codebase
    code_docs = [
        Node(
            text="""
            class UserService:
                def __init__(self, db, cache):
                    self.db = db
                    self.cache = cache
                
                async def create_user(self, email: str, name: str) -> User:
                    # Validate input
                    if not email or '@' not in email:
                        raise ValueError("Invalid email")
                    
                    # Check if exists
                    if await self.db.user_exists(email):
                        raise UserExistsError(email)
                    
                    # Create user
                    user = User(
                        id=str(uuid4()),
                        email=email,
                        name=name,
                        created_at=datetime.utcnow()
                    )
                    
                    await self.db.save(user)
                    await self.cache.set(f"user:{user.id}", user)
                    
                    return user
                
                async def get_user(self, user_id: str) -> Optional[User]:
                    # Check cache first
                    cached = await self.cache.get(f"user:{user_id}")
                    if cached:
                        return cached
                    
                    # Fetch from database
                    user = await self.db.get_user(user_id)
                    if user:
                        await self.cache.set(f"user:{user.id}", user)
                    
                    return user
            """,
            metadata={"file": "services/user_service.py"},
        ),
        Node(
            text="""
            class AuthMiddleware:
                def __init__(self, secret_key: str):
                    self.secret_key = secret_key
                
                async def __call__(self, request, call_next):
                    token = request.headers.get("Authorization", "").replace("Bearer ", "")
                    
                    if not token:
                        return JSONResponse(
                            {"error": "Missing token"},
                            status_code=401
                        )
                    
                    try:
                        payload = jwt.decode(token, self.secret_key, algorithms=["HS256"])
                        request.state.user_id = payload["sub"]
                    except JWTError:
                        return JSONResponse(
                            {"error": "Invalid token"},
                            status_code=401
                        )
                    
                    return await call_next(request)
            """,
            metadata={"file": "middleware/auth.py"},
        ),
    ]
    
    # Simulated compression results
    def mock_compress(text: str) -> dict:
        """Simulate TokMan compression."""
        original_tokens = len(text.split())
        
        # Simulate aggressive compression for code
        compressed = text
        # Remove comments
        import re
        compressed = re.sub(r'#.*$', '', compressed, flags=re.MULTILINE)
        # Remove blank lines
        compressed = re.sub(r'\n\s*\n', '\n', compressed)
        # Compress whitespace
        compressed = re.sub(r'\s+', ' ', compressed)
        
        final_tokens = len(compressed.split())
        
        return {
            "output": compressed.strip(),
            "original_tokens": original_tokens,
            "final_tokens": final_tokens,
            "reduction_percent": (1 - final_tokens / original_tokens) * 100 if original_tokens > 0 else 0,
        }
    
    # Process with mock
    combined = "\n\n".join(node.text for node in code_docs)
    result = mock_compress(combined)
    
    print(f"\nOriginal: {result['original_tokens']} tokens")
    print(f"Compressed: {result['final_tokens']} tokens")
    print(f"Reduction: {result['reduction_percent']:.1f}%")
    print(f"\nCompressed output:\n{result['output'][:500]}...")


if __name__ == "__main__":
    # Run mock demo (doesn't require running server)
    demo_with_mock_server()
    
    print("\n" + "=" * 60)
    print("To use with real TokMan server:")
    print("  1. Start server: go run ./cmd/server")
    print("  2. Run: python integration.py")
    print("=" * 60)
