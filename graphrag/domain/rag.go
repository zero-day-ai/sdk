package domain

import (
	"github.com/zero-day-ai/sdk/graphrag"
)

// VectorStore represents a vector database for embeddings.
// Vector stores index and retrieve high-dimensional embeddings for semantic search.
//
// Example:
//
//	store := &VectorStore{
//	    Name:           "security-kb",
//	    Provider:       "pinecone",
//	    Type:           "dense",
//	    Dimensions:     1536,
//	    DistanceMetric: "cosine",
//	    IndexCount:     150000,
//	}
//
// Identifying Properties:
//   - name (required): Unique name of the vector store
//
// Relationships:
//   - None (root node)
//   - Children: VectorIndex nodes
type VectorStore struct {
	// Name is the unique identifier for this vector store.
	// This is an identifying property and is required.
	Name string

	// Provider is the vector database provider.
	// Optional. Examples: "pinecone", "weaviate", "qdrant", "milvus", "chroma"
	Provider string

	// Type is the type of vector index.
	// Optional. Common values: "dense", "sparse", "hybrid"
	Type string

	// Dimensions is the dimensionality of the embedding vectors.
	// Optional. Common values: 384, 768, 1536, 3072
	Dimensions int

	// DistanceMetric is the distance metric used for similarity search.
	// Optional. Common values: "cosine", "euclidean", "dot_product", "manhattan"
	DistanceMetric string

	// IndexCount is the number of vectors currently indexed.
	// Optional.
	IndexCount int
}

// NodeType returns the canonical node type for VectorStore nodes.
// Implements GraphNode interface.
func (v *VectorStore) NodeType() string {
	return graphrag.NodeTypeVectorStore
}

// IdentifyingProperties returns the properties that uniquely identify this vector store.
// For VectorStore nodes, only name is identifying.
// Implements GraphNode interface.
func (v *VectorStore) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: v.Name,
	}
}

// Properties returns all properties to set on the VectorStore node.
// Implements GraphNode interface.
func (v *VectorStore) Properties() map[string]any {
	props := v.IdentifyingProperties()

	if v.Provider != "" {
		props[graphrag.PropProvider] = v.Provider
	}
	if v.Type != "" {
		props["type"] = v.Type
	}
	if v.Dimensions > 0 {
		props[graphrag.PropDimensions] = v.Dimensions
	}
	if v.DistanceMetric != "" {
		props[graphrag.PropDistanceMetric] = v.DistanceMetric
	}
	if v.IndexCount > 0 {
		props["index_count"] = v.IndexCount
	}

	return props
}

// ParentRef returns nil because VectorStore is a root node with no parent.
// Implements GraphNode interface.
func (v *VectorStore) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because VectorStore is a root node.
// Implements GraphNode interface.
func (v *VectorStore) RelationshipType() string {
	return ""
}

// VectorIndex represents an index within a vector store.
// Indexes organize vectors for efficient similarity search with specific configurations.
//
// Example:
//
//	index := &VectorIndex{
//	    StoreID:    "security-kb",
//	    Name:       "findings-index",
//	    Dimensions: 1536,
//	    Algorithm:  "hnsw",
//	    Config:     map[string]any{"m": 16, "ef_construction": 200},
//	}
//
// Identifying Properties:
//   - store_id (required): Parent vector store name
//   - name (required): Index name within the store
//
// Relationships:
//   - Parent: VectorStore node (via INDEXED_IN relationship)
type VectorIndex struct {
	// StoreID is the identifier of the parent vector store.
	// This is an identifying property and is required.
	StoreID string

	// Name is the unique name of this index within the store.
	// This is an identifying property and is required.
	Name string

	// Dimensions is the dimensionality of vectors in this index.
	// Optional.
	Dimensions int

	// Algorithm is the indexing algorithm used.
	// Optional. Examples: "hnsw", "ivf", "flat", "annoy"
	Algorithm string

	// Config contains algorithm-specific configuration parameters.
	// Optional. Example: {"m": 16, "ef_construction": 200, "ef_search": 100}
	Config map[string]any
}

// NodeType returns the canonical node type for VectorIndex nodes.
// Implements GraphNode interface.
func (vi *VectorIndex) NodeType() string {
	return graphrag.NodeTypeVectorIndex
}

// IdentifyingProperties returns the properties that uniquely identify this vector index.
// For VectorIndex nodes, store_id and name are both identifying.
// Implements GraphNode interface.
func (vi *VectorIndex) IdentifyingProperties() map[string]any {
	return map[string]any{
		"store_id":        vi.StoreID,
		graphrag.PropName: vi.Name,
	}
}

// Properties returns all properties to set on the VectorIndex node.
// Implements GraphNode interface.
func (vi *VectorIndex) Properties() map[string]any {
	props := vi.IdentifyingProperties()

	if vi.Dimensions > 0 {
		props[graphrag.PropDimensions] = vi.Dimensions
	}
	if vi.Algorithm != "" {
		props["algorithm"] = vi.Algorithm
	}
	if len(vi.Config) > 0 {
		props["config"] = vi.Config
	}

	return props
}

// ParentRef returns a reference to the parent VectorStore node.
// Implements GraphNode interface.
func (vi *VectorIndex) ParentRef() *NodeRef {
	if vi.StoreID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeVectorStore,
		Properties: map[string]any{
			graphrag.PropName: vi.StoreID,
		},
	}
}

// RelationshipType returns the relationship type to the parent VectorStore node.
// Implements GraphNode interface.
func (vi *VectorIndex) RelationshipType() string {
	return graphrag.RelTypeIndexedIn
}

// Document represents a source document for RAG.
// Documents are the primary unit of content that gets chunked and embedded.
//
// Example:
//
//	doc := &Document{
//	    Source:      "https://docs.example.com/api/v1",
//	    DocumentID:  "doc-12345",
//	    Title:       "API Documentation v1",
//	    ContentHash: "sha256:abc123...",
//	    ChunkCount:  42,
//	    Metadata:    map[string]any{"version": "1.0", "language": "en"},
//	}
//
// Identifying Properties:
//   - document_id (required): Unique document identifier
//
// Relationships:
//   - None (root node)
//   - Children: DocumentChunk nodes
type Document struct {
	// Source is the origin URL or path of the document.
	// Optional. Example: "https://docs.example.com/api/v1", "/data/corpus/file.pdf"
	Source string

	// DocumentID is the unique identifier for this document.
	// This is an identifying property and is required.
	DocumentID string

	// Title is the human-readable title of the document.
	// Optional.
	Title string

	// ContentHash is a hash of the document content for change detection.
	// Optional. Example: "sha256:abc123def456..."
	ContentHash string

	// ChunkCount is the number of chunks this document was split into.
	// Optional.
	ChunkCount int

	// Metadata contains arbitrary metadata about the document.
	// Optional. Example: {"author": "John Doe", "version": "1.0", "language": "en"}
	Metadata map[string]any
}

// NodeType returns the canonical node type for Document nodes.
// Implements GraphNode interface.
func (d *Document) NodeType() string {
	return graphrag.NodeTypeDocument
}

// IdentifyingProperties returns the properties that uniquely identify this document.
// For Document nodes, only document_id is identifying.
// Implements GraphNode interface.
func (d *Document) IdentifyingProperties() map[string]any {
	return map[string]any{
		"document_id": d.DocumentID,
	}
}

// Properties returns all properties to set on the Document node.
// Implements GraphNode interface.
func (d *Document) Properties() map[string]any {
	props := d.IdentifyingProperties()

	if d.Source != "" {
		props["source"] = d.Source
	}
	if d.Title != "" {
		props[graphrag.PropTitle] = d.Title
	}
	if d.ContentHash != "" {
		props["content_hash"] = d.ContentHash
	}
	if d.ChunkCount > 0 {
		props["chunk_count"] = d.ChunkCount
	}
	if len(d.Metadata) > 0 {
		props["metadata"] = d.Metadata
	}

	return props
}

// ParentRef returns nil because Document is a root node with no parent.
// Implements GraphNode interface.
func (d *Document) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Document is a root node.
// Implements GraphNode interface.
func (d *Document) RelationshipType() string {
	return ""
}

// DocumentChunk represents a chunk of a document.
// Documents are split into chunks for embedding and retrieval.
//
// Example:
//
//	chunk := &DocumentChunk{
//	    DocumentID:  "doc-12345",
//	    ChunkIndex:  5,
//	    Content:     "This section describes the authentication flow...",
//	    EmbeddingID: "emb-67890",
//	    Metadata:    map[string]any{"section": "authentication", "page": 12},
//	    TokenCount:  256,
//	}
//
// Identifying Properties:
//   - document_id (required): Parent document identifier
//   - chunk_index (required): Zero-based index of this chunk within the document
//
// Relationships:
//   - Parent: Document node (via CHUNKED_FROM relationship)
type DocumentChunk struct {
	// DocumentID is the identifier of the parent document.
	// This is an identifying property and is required.
	DocumentID string

	// ChunkIndex is the zero-based index of this chunk within the document.
	// This is an identifying property and is required.
	ChunkIndex int

	// Content is the text content of this chunk.
	// Optional but typically populated.
	Content string

	// EmbeddingID is the identifier of the embedding for this chunk.
	// Optional. Links to the Embedding node.
	EmbeddingID string

	// Metadata contains arbitrary metadata about the chunk.
	// Optional. Example: {"section": "intro", "page": 5, "start_line": 100}
	Metadata map[string]any

	// TokenCount is the number of tokens in this chunk.
	// Optional. Useful for tracking chunk sizes.
	TokenCount int
}

// NodeType returns the canonical node type for DocumentChunk nodes.
// Implements GraphNode interface.
func (dc *DocumentChunk) NodeType() string {
	return graphrag.NodeTypeDocumentChunk
}

// IdentifyingProperties returns the properties that uniquely identify this chunk.
// For DocumentChunk nodes, document_id and chunk_index are both identifying.
// Implements GraphNode interface.
func (dc *DocumentChunk) IdentifyingProperties() map[string]any {
	return map[string]any{
		"document_id":           dc.DocumentID,
		graphrag.PropChunkIndex: dc.ChunkIndex,
	}
}

// Properties returns all properties to set on the DocumentChunk node.
// Implements GraphNode interface.
func (dc *DocumentChunk) Properties() map[string]any {
	props := dc.IdentifyingProperties()

	if dc.Content != "" {
		props["content"] = dc.Content
	}
	if dc.EmbeddingID != "" {
		props[graphrag.PropEmbeddingID] = dc.EmbeddingID
	}
	if len(dc.Metadata) > 0 {
		props["metadata"] = dc.Metadata
	}
	if dc.TokenCount > 0 {
		props["token_count"] = dc.TokenCount
	}

	return props
}

// ParentRef returns a reference to the parent Document node.
// Implements GraphNode interface.
func (dc *DocumentChunk) ParentRef() *NodeRef {
	if dc.DocumentID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeDocument,
		Properties: map[string]any{
			"document_id": dc.DocumentID,
		},
	}
}

// RelationshipType returns the relationship type to the parent Document node.
// Implements GraphNode interface.
func (dc *DocumentChunk) RelationshipType() string {
	return graphrag.RelTypeChunkedFrom
}

// KnowledgeBase represents a collection of documents for RAG.
// Knowledge bases organize related documents into a queryable collection.
//
// Example:
//
//	kb := &KnowledgeBase{
//	    Name:           "security-documentation",
//	    Description:    "Security best practices and vulnerability documentation",
//	    DocumentCount:  1250,
//	    VectorStoreID:  "security-kb",
//	}
//
// Identifying Properties:
//   - name (required): Unique knowledge base identifier
//
// Relationships:
//   - None (root node)
type KnowledgeBase struct {
	// Name is the unique identifier for this knowledge base.
	// This is an identifying property and is required.
	Name string

	// Description describes the purpose and contents of this knowledge base.
	// Optional.
	Description string

	// DocumentCount is the number of documents in this knowledge base.
	// Optional.
	DocumentCount int

	// VectorStoreID is the identifier of the vector store backing this KB.
	// Optional. Links to the VectorStore node.
	VectorStoreID string
}

// NodeType returns the canonical node type for KnowledgeBase nodes.
// Implements GraphNode interface.
func (k *KnowledgeBase) NodeType() string {
	return graphrag.NodeTypeKnowledgeBase
}

// IdentifyingProperties returns the properties that uniquely identify this knowledge base.
// For KnowledgeBase nodes, only name is identifying.
// Implements GraphNode interface.
func (k *KnowledgeBase) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: k.Name,
	}
}

// Properties returns all properties to set on the KnowledgeBase node.
// Implements GraphNode interface.
func (k *KnowledgeBase) Properties() map[string]any {
	props := k.IdentifyingProperties()

	if k.Description != "" {
		props[graphrag.PropDescription] = k.Description
	}
	if k.DocumentCount > 0 {
		props["document_count"] = k.DocumentCount
	}
	if k.VectorStoreID != "" {
		props["vector_store_id"] = k.VectorStoreID
	}

	return props
}

// ParentRef returns nil because KnowledgeBase is a root node with no parent.
// Implements GraphNode interface.
func (k *KnowledgeBase) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because KnowledgeBase is a root node.
// Implements GraphNode interface.
func (k *KnowledgeBase) RelationshipType() string {
	return ""
}

// Retriever represents a retrieval component for document search.
// Retrievers implement different search strategies over vector stores.
//
// Example:
//
//	retriever := &Retriever{
//	    Name:          "hybrid-retriever",
//	    Type:          "hybrid",
//	    VectorStoreID: "security-kb",
//	    Config:        map[string]any{"alpha": 0.5, "sparse_weight": 0.3},
//	    TopK:          10,
//	}
//
// Identifying Properties:
//   - name (required): Unique retriever identifier
//
// Relationships:
//   - Parent: VectorStore node (implicit via vector_store_id)
type Retriever struct {
	// Name is the unique identifier for this retriever.
	// This is an identifying property and is required.
	Name string

	// Type is the retrieval strategy type.
	// Optional. Common values: "similarity", "mmr", "hybrid", "bm25"
	Type string

	// VectorStoreID is the identifier of the vector store this retriever queries.
	// Optional. Links to the VectorStore node.
	VectorStoreID string

	// Config contains retriever-specific configuration parameters.
	// Optional. Example: {"alpha": 0.5, "diversity_weight": 0.3}
	Config map[string]any

	// TopK is the number of results to retrieve.
	// Optional. Default varies by implementation, typically 5-10.
	TopK int
}

// NodeType returns the canonical node type for Retriever nodes.
// Implements GraphNode interface.
func (r *Retriever) NodeType() string {
	return graphrag.NodeTypeRetriever
}

// IdentifyingProperties returns the properties that uniquely identify this retriever.
// For Retriever nodes, only name is identifying.
// Implements GraphNode interface.
func (r *Retriever) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: r.Name,
	}
}

// Properties returns all properties to set on the Retriever node.
// Implements GraphNode interface.
func (r *Retriever) Properties() map[string]any {
	props := r.IdentifyingProperties()

	if r.Type != "" {
		props["type"] = r.Type
	}
	if r.VectorStoreID != "" {
		props["vector_store_id"] = r.VectorStoreID
	}
	if len(r.Config) > 0 {
		props["config"] = r.Config
	}
	if r.TopK > 0 {
		props["top_k"] = r.TopK
	}

	return props
}

// ParentRef returns nil as Retriever is treated as a root node.
// While it references a VectorStore, the relationship is not hierarchical.
// Implements GraphNode interface.
func (r *Retriever) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Retriever is a root node.
// Implements GraphNode interface.
func (r *Retriever) RelationshipType() string {
	return ""
}

// RAGPipeline represents a complete RAG pipeline configuration.
// Pipelines combine retrieval, generation, and optional reranking steps.
//
// Example:
//
//	pipeline := &RAGPipeline{
//	    Name:           "security-qa-pipeline",
//	    Description:    "Question-answering pipeline for security documentation",
//	    RetrieverID:    "hybrid-retriever",
//	    GeneratorModel: "gpt-4-turbo",
//	    Config:         map[string]any{"temperature": 0.0, "max_tokens": 1000},
//	}
//
// Identifying Properties:
//   - name (required): Unique pipeline identifier
//
// Relationships:
//   - None (root node)
type RAGPipeline struct {
	// Name is the unique identifier for this pipeline.
	// This is an identifying property and is required.
	Name string

	// Description describes the purpose and configuration of this pipeline.
	// Optional.
	Description string

	// RetrieverID is the identifier of the retriever component.
	// Optional. Links to the Retriever node.
	RetrieverID string

	// GeneratorModel is the LLM model used for generation.
	// Optional. Example: "gpt-4-turbo", "claude-3-opus"
	GeneratorModel string

	// Config contains pipeline-specific configuration parameters.
	// Optional. Example: {"temperature": 0.0, "rerank": true, "max_docs": 5}
	Config map[string]any
}

// NodeType returns the canonical node type for RAGPipeline nodes.
// Implements GraphNode interface.
func (r *RAGPipeline) NodeType() string {
	return graphrag.NodeTypeRAGPipeline
}

// IdentifyingProperties returns the properties that uniquely identify this pipeline.
// For RAGPipeline nodes, only name is identifying.
// Implements GraphNode interface.
func (r *RAGPipeline) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: r.Name,
	}
}

// Properties returns all properties to set on the RAGPipeline node.
// Implements GraphNode interface.
func (r *RAGPipeline) Properties() map[string]any {
	props := r.IdentifyingProperties()

	if r.Description != "" {
		props[graphrag.PropDescription] = r.Description
	}
	if r.RetrieverID != "" {
		props["retriever_id"] = r.RetrieverID
	}
	if r.GeneratorModel != "" {
		props["generator_model"] = r.GeneratorModel
	}
	if len(r.Config) > 0 {
		props["config"] = r.Config
	}

	return props
}

// ParentRef returns nil because RAGPipeline is a root node with no parent.
// Implements GraphNode interface.
func (r *RAGPipeline) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because RAGPipeline is a root node.
// Implements GraphNode interface.
func (r *RAGPipeline) RelationshipType() string {
	return ""
}

// Embedding represents an embedding vector for a document chunk.
// Embeddings are high-dimensional vector representations enabling semantic search.
// Note: The actual vector data is typically stored in the vector store, not in the graph.
//
// Example:
//
//	embedding := &Embedding{
//	    ID:          "emb-67890",
//	    DocumentID:  "doc-12345",
//	    ChunkIndex:  5,
//	    Model:       "text-embedding-3-large",
//	}
//
// Identifying Properties:
//   - id (required): Unique embedding identifier
//
// Relationships:
//   - Parent: DocumentChunk node (via EMBEDDED_AS relationship)
type Embedding struct {
	// ID is the unique identifier for this embedding.
	// This is an identifying property and is required.
	ID string

	// DocumentID is the identifier of the source document.
	// Optional but typically populated for traceability.
	DocumentID string

	// ChunkIndex is the index of the chunk this embedding represents.
	// Optional. Used to link back to the specific DocumentChunk.
	ChunkIndex int

	// Model is the embedding model used to generate this vector.
	// Optional. Example: "text-embedding-3-large", "voyage-2"
	Model string

	// Note: The actual vector data is not stored here - it lives in the vector store.
	// This node tracks metadata about the embedding for graph traversal.
}

// NodeType returns the canonical node type for Embedding nodes.
// Implements GraphNode interface.
func (e *Embedding) NodeType() string {
	return graphrag.NodeTypeEmbedding
}

// IdentifyingProperties returns the properties that uniquely identify this embedding.
// For Embedding nodes, only id is identifying.
// Implements GraphNode interface.
func (e *Embedding) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": e.ID,
	}
}

// Properties returns all properties to set on the Embedding node.
// Implements GraphNode interface.
func (e *Embedding) Properties() map[string]any {
	props := e.IdentifyingProperties()

	if e.DocumentID != "" {
		props["document_id"] = e.DocumentID
	}
	if e.ChunkIndex >= 0 {
		props[graphrag.PropChunkIndex] = e.ChunkIndex
	}
	if e.Model != "" {
		props[graphrag.PropModel] = e.Model
	}

	return props
}

// ParentRef returns a reference to the parent DocumentChunk node.
// Implements GraphNode interface.
func (e *Embedding) ParentRef() *NodeRef {
	if e.DocumentID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeDocumentChunk,
		Properties: map[string]any{
			"document_id":           e.DocumentID,
			graphrag.PropChunkIndex: e.ChunkIndex,
		},
	}
}

// RelationshipType returns the relationship type to the parent DocumentChunk node.
// Implements GraphNode interface.
func (e *Embedding) RelationshipType() string {
	return graphrag.RelTypeEmbeddedAs
}

// Reranker represents a reranking model for retrieval results.
// Rerankers improve retrieval quality by reordering results with more sophisticated models.
//
// Example:
//
//	reranker := &Reranker{
//	    Name:     "cross-encoder-reranker",
//	    Model:    "cross-encoder/ms-marco-MiniLM-L-12-v2",
//	    Provider: "sentence-transformers",
//	    Config:   map[string]any{"batch_size": 32, "top_k": 10},
//	}
//
// Identifying Properties:
//   - name (required): Unique reranker identifier
//
// Relationships:
//   - None (root node)
type Reranker struct {
	// Name is the unique identifier for this reranker.
	// This is an identifying property and is required.
	Name string

	// Model is the reranking model identifier.
	// Optional. Example: "cross-encoder/ms-marco-MiniLM-L-12-v2"
	Model string

	// Provider is the model provider or framework.
	// Optional. Examples: "sentence-transformers", "cohere", "openai"
	Provider string

	// Config contains reranker-specific configuration parameters.
	// Optional. Example: {"batch_size": 32, "top_k": 10}
	Config map[string]any
}

// NodeType returns the canonical node type for Reranker nodes.
// Implements GraphNode interface.
func (r *Reranker) NodeType() string {
	return graphrag.NodeTypeReranker
}

// IdentifyingProperties returns the properties that uniquely identify this reranker.
// For Reranker nodes, only name is identifying.
// Implements GraphNode interface.
func (r *Reranker) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: r.Name,
	}
}

// Properties returns all properties to set on the Reranker node.
// Implements GraphNode interface.
func (r *Reranker) Properties() map[string]any {
	props := r.IdentifyingProperties()

	if r.Model != "" {
		props[graphrag.PropModel] = r.Model
	}
	if r.Provider != "" {
		props[graphrag.PropProvider] = r.Provider
	}
	if len(r.Config) > 0 {
		props["config"] = r.Config
	}

	return props
}

// ParentRef returns nil because Reranker is a root node with no parent.
// Implements GraphNode interface.
func (r *Reranker) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Reranker is a root node.
// Implements GraphNode interface.
func (r *Reranker) RelationshipType() string {
	return ""
}

// ChunkingStrategy represents a document chunking strategy.
// Chunking strategies define how documents are split into retrievable units.
//
// Example:
//
//	strategy := &ChunkingStrategy{
//	    Name:      "semantic-chunking",
//	    Type:      "semantic",
//	    ChunkSize: 512,
//	    Overlap:   50,
//	    Separator: "\n\n",
//	}
//
// Identifying Properties:
//   - name (required): Unique strategy identifier
//
// Relationships:
//   - None (root node)
type ChunkingStrategy struct {
	// Name is the unique identifier for this chunking strategy.
	// This is an identifying property and is required.
	Name string

	// Type is the chunking algorithm type.
	// Optional. Common values: "fixed", "semantic", "recursive", "sentence", "paragraph"
	Type string

	// ChunkSize is the target size of each chunk in tokens or characters.
	// Optional. Example: 512, 1024
	ChunkSize int

	// Overlap is the number of tokens/characters to overlap between chunks.
	// Optional. Example: 50, 100
	Overlap int

	// Separator is the text separator used for splitting.
	// Optional. Examples: "\n\n", "\n", ".", " "
	Separator string
}

// NodeType returns the canonical node type for ChunkingStrategy nodes.
// Implements GraphNode interface.
func (c *ChunkingStrategy) NodeType() string {
	return graphrag.NodeTypeChunkingStrategy
}

// IdentifyingProperties returns the properties that uniquely identify this strategy.
// For ChunkingStrategy nodes, only name is identifying.
// Implements GraphNode interface.
func (c *ChunkingStrategy) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: c.Name,
	}
}

// Properties returns all properties to set on the ChunkingStrategy node.
// Implements GraphNode interface.
func (c *ChunkingStrategy) Properties() map[string]any {
	props := c.IdentifyingProperties()

	if c.Type != "" {
		props["type"] = c.Type
	}
	if c.ChunkSize > 0 {
		props["chunk_size"] = c.ChunkSize
	}
	if c.Overlap > 0 {
		props["overlap"] = c.Overlap
	}
	if c.Separator != "" {
		props["separator"] = c.Separator
	}

	return props
}

// ParentRef returns nil because ChunkingStrategy is a root node with no parent.
// Implements GraphNode interface.
func (c *ChunkingStrategy) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because ChunkingStrategy is a root node.
// Implements GraphNode interface.
func (c *ChunkingStrategy) RelationshipType() string {
	return ""
}

// RetrievalResult represents a result from a retrieval query.
// Results track which chunks were retrieved for a specific query.
//
// Example:
//
//	result := &RetrievalResult{
//	    QueryID:  "query-abc123",
//	    ChunkID:  "doc-12345-chunk-5",
//	    Score:    0.87,
//	    Rank:     3,
//	    Metadata: map[string]any{"reranked": true, "original_rank": 7},
//	}
//
// Identifying Properties:
//   - query_id (required): Identifier of the query
//   - chunk_id (required): Identifier of the retrieved chunk
//
// Relationships:
//   - None (root node representing query results)
type RetrievalResult struct {
	// QueryID is the identifier of the query that produced this result.
	// This is an identifying property and is required.
	QueryID string

	// ChunkID is the identifier of the retrieved document chunk.
	// This is an identifying property and is required.
	ChunkID string

	// Score is the similarity or relevance score.
	// Optional. Typically a float between 0.0 and 1.0.
	Score float64

	// Rank is the position of this result in the ranked list.
	// Optional. 1-indexed rank (1 = top result).
	Rank int

	// Metadata contains additional information about the retrieval.
	// Optional. Example: {"reranked": true, "original_rank": 7}
	Metadata map[string]any
}

// NodeType returns the canonical node type for RetrievalResult nodes.
// Implements GraphNode interface.
func (r *RetrievalResult) NodeType() string {
	return graphrag.NodeTypeRetrievalResult
}

// IdentifyingProperties returns the properties that uniquely identify this result.
// For RetrievalResult nodes, query_id and chunk_id are both identifying.
// Implements GraphNode interface.
func (r *RetrievalResult) IdentifyingProperties() map[string]any {
	return map[string]any{
		"query_id": r.QueryID,
		"chunk_id": r.ChunkID,
	}
}

// Properties returns all properties to set on the RetrievalResult node.
// Implements GraphNode interface.
func (r *RetrievalResult) Properties() map[string]any {
	props := r.IdentifyingProperties()

	if r.Score > 0 {
		props["score"] = r.Score
	}
	if r.Rank > 0 {
		props["rank"] = r.Rank
	}
	if len(r.Metadata) > 0 {
		props["metadata"] = r.Metadata
	}

	return props
}

// ParentRef returns nil because RetrievalResult is a root node.
// Results represent query outcomes rather than hierarchical data.
// Implements GraphNode interface.
func (r *RetrievalResult) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because RetrievalResult is a root node.
// Implements GraphNode interface.
func (r *RetrievalResult) RelationshipType() string {
	return ""
}
