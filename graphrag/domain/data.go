package domain

import "github.com/zero-day-ai/sdk/graphrag"

// Database represents a database instance.
// Databases store structured or unstructured data.
//
// Example:
//
//	db := &Database{
//	    Name:    "production-db",
//	    Host:    "db.example.com",
//	    Type:    "postgresql",
//	    Version: "14.5",
//	}
//
// Identifying Properties:
//   - name (required): Database name
//   - host (required): Database host
//
// Relationships:
//   - None (root node)
//   - Children: Table, View, StoredProcedure nodes
type Database struct {
	// Name is the database name.
	// This is an identifying property and is required.
	Name string

	// Host is the database host or connection string.
	// This is an identifying property and is required.
	// Example: "db.example.com", "localhost:5432"
	Host string

	// Type is the database type/engine.
	// Optional. Common values: "postgresql", "mysql", "mongodb", "redis", "sqlserver"
	Type string

	// Version is the database version.
	// Optional. Example: "14.5", "8.0.31"
	Version string

	// Port is the database port.
	// Optional.
	Port int

	// Description is a description of the database.
	// Optional.
	Description string
}

func (d *Database) NodeType() string { return "database" }

func (d *Database) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: d.Name,
		"host":            d.Host,
	}
}

func (d *Database) Properties() map[string]any {
	props := d.IdentifyingProperties()
	if d.Type != "" {
		props["type"] = d.Type
	}
	if d.Version != "" {
		props["version"] = d.Version
	}
	if d.Port > 0 {
		props[graphrag.PropPort] = d.Port
	}
	if d.Description != "" {
		props[graphrag.PropDescription] = d.Description
	}
	return props
}

func (d *Database) ParentRef() *NodeRef      { return nil }
func (d *Database) RelationshipType() string { return "" }

// Table represents a database table.
// Tables store rows of structured data.
//
// Example:
//
//	table := &Table{
//	    DatabaseID: "production-db:db.example.com",
//	    Name:       "users",
//	    Schema:     "public",
//	    RowCount:   150000,
//	}
//
// Identifying Properties:
//   - database_id (required): The database this table belongs to
//   - name (required): Table name
//
// Relationships:
//   - Parent: Database node (via HAS_TABLE relationship)
//   - Children: Column, Index, Trigger nodes
type Table struct {
	// DatabaseID is the identifier of the parent database.
	// This is an identifying property and is required.
	DatabaseID string

	// Name is the table name.
	// This is an identifying property and is required.
	Name string

	// Schema is the database schema name.
	// Optional. Example: "public", "dbo"
	Schema string

	// RowCount is the approximate number of rows.
	// Optional.
	RowCount int64

	// SizeBytes is the table size in bytes.
	// Optional.
	SizeBytes int64

	// Description is a description of the table.
	// Optional.
	Description string
}

func (t *Table) NodeType() string { return "table" }

func (t *Table) IdentifyingProperties() map[string]any {
	return map[string]any{
		"database_id":     t.DatabaseID,
		graphrag.PropName: t.Name,
	}
}

func (t *Table) Properties() map[string]any {
	props := t.IdentifyingProperties()
	if t.Schema != "" {
		props["schema"] = t.Schema
	}
	if t.RowCount > 0 {
		props["row_count"] = t.RowCount
	}
	if t.SizeBytes > 0 {
		props["size_bytes"] = t.SizeBytes
	}
	if t.Description != "" {
		props[graphrag.PropDescription] = t.Description
	}
	return props
}

func (t *Table) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "database",
		Properties: map[string]any{
			"id": t.DatabaseID,
		},
	}
}

func (t *Table) RelationshipType() string { return "HAS_TABLE" }

// Column represents a column in a database table.
// Columns define the structure and data types of table data.
//
// Example:
//
//	col := &Column{
//	    TableID:  "users",
//	    Name:     "email",
//	    DataType: "varchar(255)",
//	    Nullable: false,
//	}
//
// Identifying Properties:
//   - table_id (required): The table this column belongs to
//   - name (required): Column name
//
// Relationships:
//   - Parent: Table node (via HAS_COLUMN relationship)
type Column struct {
	// TableID is the identifier of the parent table.
	// This is an identifying property and is required.
	TableID string

	// Name is the column name.
	// This is an identifying property and is required.
	Name string

	// DataType is the column data type.
	// Optional. Example: "varchar(255)", "integer", "timestamp"
	DataType string

	// Nullable indicates if the column can contain NULL values.
	// Optional. Default: true
	Nullable bool

	// DefaultValue is the default value for the column.
	// Optional.
	DefaultValue string

	// IsPrimaryKey indicates if this column is part of the primary key.
	// Optional. Default: false
	IsPrimaryKey bool

	// Description is a description of the column.
	// Optional.
	Description string
}

func (c *Column) NodeType() string { return "column" }

func (c *Column) IdentifyingProperties() map[string]any {
	return map[string]any{
		"table_id":        c.TableID,
		graphrag.PropName: c.Name,
	}
}

func (c *Column) Properties() map[string]any {
	props := c.IdentifyingProperties()
	if c.DataType != "" {
		props["data_type"] = c.DataType
	}
	props["nullable"] = c.Nullable
	if c.DefaultValue != "" {
		props["default_value"] = c.DefaultValue
	}
	props["is_primary_key"] = c.IsPrimaryKey
	if c.Description != "" {
		props[graphrag.PropDescription] = c.Description
	}
	return props
}

func (c *Column) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "table",
		Properties: map[string]any{
			"id": c.TableID,
		},
	}
}

func (c *Column) RelationshipType() string { return "HAS_COLUMN" }

// Index represents a database index.
// Indexes improve query performance by organizing data.
//
// Example:
//
//	idx := &Index{
//	    TableID: "users",
//	    Name:    "idx_users_email",
//	    Type:    "btree",
//	    Unique:  true,
//	}
//
// Identifying Properties:
//   - table_id (required): The table this index belongs to
//   - name (required): Index name
//
// Relationships:
//   - Parent: Table node (via HAS_INDEX relationship)
type Index struct {
	// TableID is the identifier of the parent table.
	// This is an identifying property and is required.
	TableID string

	// Name is the index name.
	// This is an identifying property and is required.
	Name string

	// Type is the index type.
	// Optional. Common values: "btree", "hash", "gin", "gist"
	Type string

	// Unique indicates if this is a unique index.
	// Optional. Default: false
	Unique bool

	// Columns are the columns included in the index.
	// Optional. Example: ["email", "last_name"]
	Columns []string

	// Description is a description of the index.
	// Optional.
	Description string
}

func (i *Index) NodeType() string { return "index" }

func (i *Index) IdentifyingProperties() map[string]any {
	return map[string]any{
		"table_id":        i.TableID,
		graphrag.PropName: i.Name,
	}
}

func (i *Index) Properties() map[string]any {
	props := i.IdentifyingProperties()
	if i.Type != "" {
		props["type"] = i.Type
	}
	props["unique"] = i.Unique
	if len(i.Columns) > 0 {
		props["columns"] = i.Columns
	}
	if i.Description != "" {
		props[graphrag.PropDescription] = i.Description
	}
	return props
}

func (i *Index) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "table",
		Properties: map[string]any{
			"id": i.TableID,
		},
	}
}

func (i *Index) RelationshipType() string { return "HAS_INDEX" }

// View represents a database view.
// Views are virtual tables based on SQL queries.
//
// Example:
//
//	view := &View{
//	    DatabaseID: "production-db:db.example.com",
//	    Name:       "active_users",
//	    Schema:     "public",
//	}
//
// Identifying Properties:
//   - database_id (required): The database this view belongs to
//   - name (required): View name
//
// Relationships:
//   - Parent: Database node (via HAS_VIEW relationship)
type View struct {
	// DatabaseID is the identifier of the parent database.
	// This is an identifying property and is required.
	DatabaseID string

	// Name is the view name.
	// This is an identifying property and is required.
	Name string

	// Schema is the database schema name.
	// Optional. Example: "public", "dbo"
	Schema string

	// Definition is the SQL definition of the view.
	// Optional.
	Definition string

	// Description is a description of the view.
	// Optional.
	Description string
}

func (v *View) NodeType() string { return "view" }

func (v *View) IdentifyingProperties() map[string]any {
	return map[string]any{
		"database_id":     v.DatabaseID,
		graphrag.PropName: v.Name,
	}
}

func (v *View) Properties() map[string]any {
	props := v.IdentifyingProperties()
	if v.Schema != "" {
		props["schema"] = v.Schema
	}
	if v.Definition != "" {
		props["definition"] = v.Definition
	}
	if v.Description != "" {
		props[graphrag.PropDescription] = v.Description
	}
	return props
}

func (v *View) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "database",
		Properties: map[string]any{
			"id": v.DatabaseID,
		},
	}
}

func (v *View) RelationshipType() string { return "HAS_VIEW" }

// StoredProcedure represents a database stored procedure.
// Stored procedures are precompiled SQL code that can be executed.
//
// Example:
//
//	sp := &StoredProcedure{
//	    DatabaseID: "production-db:db.example.com",
//	    Name:       "sp_create_user",
//	    Schema:     "public",
//	}
//
// Identifying Properties:
//   - database_id (required): The database this stored procedure belongs to
//   - name (required): Stored procedure name
//
// Relationships:
//   - Parent: Database node (via HAS_STORED_PROCEDURE relationship)
type StoredProcedure struct {
	// DatabaseID is the identifier of the parent database.
	// This is an identifying property and is required.
	DatabaseID string

	// Name is the stored procedure name.
	// This is an identifying property and is required.
	Name string

	// Schema is the database schema name.
	// Optional. Example: "public", "dbo"
	Schema string

	// Language is the programming language used.
	// Optional. Common values: "plpgsql", "tsql", "plsql"
	Language string

	// Definition is the source code of the stored procedure.
	// Optional.
	Definition string

	// Description is a description of the stored procedure.
	// Optional.
	Description string
}

func (s *StoredProcedure) NodeType() string { return "stored_procedure" }

func (s *StoredProcedure) IdentifyingProperties() map[string]any {
	return map[string]any{
		"database_id":     s.DatabaseID,
		graphrag.PropName: s.Name,
	}
}

func (s *StoredProcedure) Properties() map[string]any {
	props := s.IdentifyingProperties()
	if s.Schema != "" {
		props["schema"] = s.Schema
	}
	if s.Language != "" {
		props["language"] = s.Language
	}
	if s.Definition != "" {
		props["definition"] = s.Definition
	}
	if s.Description != "" {
		props[graphrag.PropDescription] = s.Description
	}
	return props
}

func (s *StoredProcedure) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "database",
		Properties: map[string]any{
			"id": s.DatabaseID,
		},
	}
}

func (s *StoredProcedure) RelationshipType() string { return "HAS_STORED_PROCEDURE" }

// Trigger represents a database trigger.
// Triggers automatically execute code in response to database events.
//
// Example:
//
//	trigger := &Trigger{
//	    TableID: "users",
//	    Name:    "trg_users_audit",
//	    Event:   "INSERT",
//	    Timing:  "AFTER",
//	}
//
// Identifying Properties:
//   - table_id (required): The table this trigger belongs to
//   - name (required): Trigger name
//
// Relationships:
//   - Parent: Table node (via HAS_TRIGGER relationship)
type Trigger struct {
	// TableID is the identifier of the parent table.
	// This is an identifying property and is required.
	TableID string

	// Name is the trigger name.
	// This is an identifying property and is required.
	Name string

	// Event is the event that fires the trigger.
	// Optional. Common values: "INSERT", "UPDATE", "DELETE"
	Event string

	// Timing is when the trigger fires.
	// Optional. Common values: "BEFORE", "AFTER", "INSTEAD OF"
	Timing string

	// Definition is the trigger code/definition.
	// Optional.
	Definition string

	// Enabled indicates if the trigger is enabled.
	// Optional. Default: true
	Enabled bool
}

func (t *Trigger) NodeType() string { return "trigger" }

func (t *Trigger) IdentifyingProperties() map[string]any {
	return map[string]any{
		"table_id":        t.TableID,
		graphrag.PropName: t.Name,
	}
}

func (t *Trigger) Properties() map[string]any {
	props := t.IdentifyingProperties()
	if t.Event != "" {
		props["event"] = t.Event
	}
	if t.Timing != "" {
		props["timing"] = t.Timing
	}
	if t.Definition != "" {
		props["definition"] = t.Definition
	}
	props["enabled"] = t.Enabled
	return props
}

func (t *Trigger) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "table",
		Properties: map[string]any{
			"id": t.TableID,
		},
	}
}

func (t *Trigger) RelationshipType() string { return "HAS_TRIGGER" }

// File represents a file in a filesystem or storage system.
// Files contain data or code stored on disk or in the cloud.
//
// Example:
//
//	file := &File{
//	    Path:     "/etc/nginx/nginx.conf",
//	    Name:     "nginx.conf",
//	    Type:     "config",
//	    SizeBytes: 4096,
//	}
//
// Identifying Properties:
//   - path (required): Full file path
//
// Relationships:
//   - None (root node)
type File struct {
	// Path is the full file path.
	// This is an identifying property and is required.
	// Example: "/etc/nginx/nginx.conf", "C:\\Windows\\System32\\config"
	Path string

	// Name is the file name.
	// Optional. Example: "nginx.conf"
	Name string

	// Type is the file type.
	// Optional. Common values: "config", "log", "binary", "script", "data"
	Type string

	// Extension is the file extension.
	// Optional. Example: ".conf", ".log", ".exe"
	Extension string

	// SizeBytes is the file size in bytes.
	// Optional.
	SizeBytes int64

	// Permissions is the file permissions.
	// Optional. Example: "0644", "rwxr-xr-x"
	Permissions string

	// Owner is the file owner.
	// Optional. Example: "root", "nginx"
	Owner string

	// Hash is a cryptographic hash of the file.
	// Optional. Example: "sha256:abc123..."
	Hash string
}

func (f *File) NodeType() string { return "file" }

func (f *File) IdentifyingProperties() map[string]any {
	return map[string]any{
		"path": f.Path,
	}
}

func (f *File) Properties() map[string]any {
	props := f.IdentifyingProperties()
	if f.Name != "" {
		props[graphrag.PropName] = f.Name
	}
	if f.Type != "" {
		props["type"] = f.Type
	}
	if f.Extension != "" {
		props["extension"] = f.Extension
	}
	if f.SizeBytes > 0 {
		props["size_bytes"] = f.SizeBytes
	}
	if f.Permissions != "" {
		props["permissions"] = f.Permissions
	}
	if f.Owner != "" {
		props["owner"] = f.Owner
	}
	if f.Hash != "" {
		props["hash"] = f.Hash
	}
	return props
}

func (f *File) ParentRef() *NodeRef      { return nil }
func (f *File) RelationshipType() string { return "" }

// StorageBucket represents a cloud storage bucket.
// Buckets are containers for storing objects in cloud storage.
//
// Example:
//
//	bucket := &StorageBucket{
//	    Name:     "my-app-assets",
//	    Provider: "aws",
//	    Region:   "us-east-1",
//	}
//
// Identifying Properties:
//   - name (required): Bucket name
//   - provider (required): Cloud provider
//
// Relationships:
//   - None (root node)
//   - Children: Object nodes
type StorageBucket struct {
	// Name is the bucket name.
	// This is an identifying property and is required.
	Name string

	// Provider is the cloud storage provider.
	// This is an identifying property and is required.
	// Common values: "aws", "gcp", "azure", "minio"
	Provider string

	// Region is the cloud region where the bucket is located.
	// Optional. Example: "us-east-1", "eu-west-1"
	Region string

	// Public indicates if the bucket is publicly accessible.
	// Optional. Default: false
	Public bool

	// Versioning indicates if versioning is enabled.
	// Optional. Default: false
	Versioning bool

	// Encryption indicates if encryption is enabled.
	// Optional. Default: false
	Encryption bool

	// Description is a description of the bucket.
	// Optional.
	Description string
}

func (s *StorageBucket) NodeType() string { return "storage_bucket" }

func (s *StorageBucket) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: s.Name,
		"provider":        s.Provider,
	}
}

func (s *StorageBucket) Properties() map[string]any {
	props := s.IdentifyingProperties()
	if s.Region != "" {
		props["region"] = s.Region
	}
	props["public"] = s.Public
	props["versioning"] = s.Versioning
	props["encryption"] = s.Encryption
	if s.Description != "" {
		props[graphrag.PropDescription] = s.Description
	}
	return props
}

func (s *StorageBucket) ParentRef() *NodeRef      { return nil }
func (s *StorageBucket) RelationshipType() string { return "" }

// Object represents an object stored in a cloud storage bucket.
// Objects are files or blobs stored in object storage.
//
// Example:
//
//	obj := &Object{
//	    BucketID:  "my-app-assets:aws",
//	    Key:       "images/logo.png",
//	    SizeBytes: 15360,
//	}
//
// Identifying Properties:
//   - bucket_id (required): The bucket this object belongs to
//   - key (required): Object key/path
//
// Relationships:
//   - Parent: StorageBucket node (via CONTAINS_OBJECT relationship)
type Object struct {
	// BucketID is the identifier of the parent bucket.
	// This is an identifying property and is required.
	BucketID string

	// Key is the object key/path within the bucket.
	// This is an identifying property and is required.
	// Example: "images/logo.png", "data/2024/01/file.json"
	Key string

	// SizeBytes is the object size in bytes.
	// Optional.
	SizeBytes int64

	// ContentType is the MIME content type.
	// Optional. Example: "image/png", "application/json"
	ContentType string

	// ETag is the entity tag (version identifier).
	// Optional.
	ETag string

	// LastModified is when the object was last modified.
	// Optional. Unix timestamp.
	LastModified int64

	// StorageClass is the storage class.
	// Optional. Example: "STANDARD", "GLACIER", "STANDARD_IA"
	StorageClass string
}

func (o *Object) NodeType() string { return "object" }

func (o *Object) IdentifyingProperties() map[string]any {
	return map[string]any{
		"bucket_id": o.BucketID,
		"key":       o.Key,
	}
}

func (o *Object) Properties() map[string]any {
	props := o.IdentifyingProperties()
	if o.SizeBytes > 0 {
		props["size_bytes"] = o.SizeBytes
	}
	if o.ContentType != "" {
		props["content_type"] = o.ContentType
	}
	if o.ETag != "" {
		props["etag"] = o.ETag
	}
	if o.LastModified > 0 {
		props["last_modified"] = o.LastModified
	}
	if o.StorageClass != "" {
		props["storage_class"] = o.StorageClass
	}
	return props
}

func (o *Object) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "storage_bucket",
		Properties: map[string]any{
			"id": o.BucketID,
		},
	}
}

func (o *Object) RelationshipType() string { return "CONTAINS_OBJECT" }

// Queue represents a message queue.
// Queues enable asynchronous communication between services.
//
// Example:
//
//	queue := &Queue{
//	    Name:     "order-processing",
//	    Provider: "aws",
//	    Type:     "fifo",
//	}
//
// Identifying Properties:
//   - name (required): Queue name
//   - provider (required): Queue provider
//
// Relationships:
//   - None (root node)
type Queue struct {
	// Name is the queue name.
	// This is an identifying property and is required.
	Name string

	// Provider is the queue provider.
	// This is an identifying property and is required.
	// Common values: "aws", "gcp", "azure", "rabbitmq", "kafka"
	Provider string

	// Type is the queue type.
	// Optional. Common values: "standard", "fifo", "priority"
	Type string

	// MessageCount is the approximate number of messages.
	// Optional.
	MessageCount int64

	// MaxMessageSize is the maximum message size in bytes.
	// Optional.
	MaxMessageSize int

	// RetentionPeriod is the message retention period in seconds.
	// Optional.
	RetentionPeriod int

	// Description is a description of the queue.
	// Optional.
	Description string
}

func (q *Queue) NodeType() string { return "queue" }

func (q *Queue) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: q.Name,
		"provider":        q.Provider,
	}
}

func (q *Queue) Properties() map[string]any {
	props := q.IdentifyingProperties()
	if q.Type != "" {
		props["type"] = q.Type
	}
	if q.MessageCount > 0 {
		props["message_count"] = q.MessageCount
	}
	if q.MaxMessageSize > 0 {
		props["max_message_size"] = q.MaxMessageSize
	}
	if q.RetentionPeriod > 0 {
		props["retention_period"] = q.RetentionPeriod
	}
	if q.Description != "" {
		props[graphrag.PropDescription] = q.Description
	}
	return props
}

func (q *Queue) ParentRef() *NodeRef      { return nil }
func (q *Queue) RelationshipType() string { return "" }

// Topic represents a pub/sub topic.
// Topics enable publish-subscribe messaging patterns.
//
// Example:
//
//	topic := &Topic{
//	    Name:     "user-events",
//	    Provider: "gcp",
//	}
//
// Identifying Properties:
//   - name (required): Topic name
//   - provider (required): Pub/sub provider
//
// Relationships:
//   - None (root node)
type Topic struct {
	// Name is the topic name.
	// This is an identifying property and is required.
	Name string

	// Provider is the pub/sub provider.
	// This is an identifying property and is required.
	// Common values: "gcp", "aws", "azure", "kafka", "pulsar"
	Provider string

	// SubscriberCount is the number of subscriptions.
	// Optional.
	SubscriberCount int

	// MessageRetention is the message retention period in seconds.
	// Optional.
	MessageRetention int

	// Description is a description of the topic.
	// Optional.
	Description string
}

func (t *Topic) NodeType() string { return "topic" }

func (t *Topic) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: t.Name,
		"provider":        t.Provider,
	}
}

func (t *Topic) Properties() map[string]any {
	props := t.IdentifyingProperties()
	if t.SubscriberCount > 0 {
		props["subscriber_count"] = t.SubscriberCount
	}
	if t.MessageRetention > 0 {
		props["message_retention"] = t.MessageRetention
	}
	if t.Description != "" {
		props[graphrag.PropDescription] = t.Description
	}
	return props
}

func (t *Topic) ParentRef() *NodeRef      { return nil }
func (t *Topic) RelationshipType() string { return "" }

// Stream represents a data stream.
// Streams process continuous flows of data in real-time.
//
// Example:
//
//	stream := &Stream{
//	    Name:      "clickstream",
//	    Type:      "kinesis",
//	    ShardCount: 4,
//	}
//
// Identifying Properties:
//   - name (required): Stream name
//
// Relationships:
//   - None (root node)
type Stream struct {
	// Name is the stream name.
	// This is an identifying property and is required.
	Name string

	// Type is the stream type.
	// Optional. Common values: "kinesis", "kafka", "pubsub", "eventhub"
	Type string

	// ShardCount is the number of shards or partitions.
	// Optional.
	ShardCount int

	// RetentionHours is the data retention period in hours.
	// Optional.
	RetentionHours int

	// ThroughputMBps is the throughput in MB/s.
	// Optional.
	ThroughputMBps int

	// Description is a description of the stream.
	// Optional.
	Description string
}

func (s *Stream) NodeType() string { return "stream" }

func (s *Stream) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: s.Name,
	}
}

func (s *Stream) Properties() map[string]any {
	props := s.IdentifyingProperties()
	if s.Type != "" {
		props["type"] = s.Type
	}
	if s.ShardCount > 0 {
		props["shard_count"] = s.ShardCount
	}
	if s.RetentionHours > 0 {
		props["retention_hours"] = s.RetentionHours
	}
	if s.ThroughputMBps > 0 {
		props["throughput_mbps"] = s.ThroughputMBps
	}
	if s.Description != "" {
		props[graphrag.PropDescription] = s.Description
	}
	return props
}

func (s *Stream) ParentRef() *NodeRef      { return nil }
func (s *Stream) RelationshipType() string { return "" }

// Cache represents a caching layer.
// Caches store frequently accessed data in memory for fast retrieval.
//
// Example:
//
//	cache := &Cache{
//	    Name:     "session-cache",
//	    Type:     "redis",
//	    Host:     "redis.example.com",
//	}
//
// Identifying Properties:
//   - name (required): Cache name
//
// Relationships:
//   - None (root node)
type Cache struct {
	// Name is the cache name.
	// This is an identifying property and is required.
	Name string

	// Type is the cache type.
	// Optional. Common values: "redis", "memcached", "elasticache", "cloudflare"
	Type string

	// Host is the cache host.
	// Optional. Example: "redis.example.com", "localhost:6379"
	Host string

	// Port is the cache port.
	// Optional.
	Port int

	// TTL is the default time-to-live in seconds.
	// Optional.
	TTL int

	// MaxSize is the maximum cache size in bytes.
	// Optional.
	MaxSize int64

	// Description is a description of the cache.
	// Optional.
	Description string
}

func (c *Cache) NodeType() string { return "cache" }

func (c *Cache) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: c.Name,
	}
}

func (c *Cache) Properties() map[string]any {
	props := c.IdentifyingProperties()
	if c.Type != "" {
		props["type"] = c.Type
	}
	if c.Host != "" {
		props["host"] = c.Host
	}
	if c.Port > 0 {
		props[graphrag.PropPort] = c.Port
	}
	if c.TTL > 0 {
		props["ttl"] = c.TTL
	}
	if c.MaxSize > 0 {
		props["max_size"] = c.MaxSize
	}
	if c.Description != "" {
		props[graphrag.PropDescription] = c.Description
	}
	return props
}

func (c *Cache) ParentRef() *NodeRef      { return nil }
func (c *Cache) RelationshipType() string { return "" }

// Schema represents a database schema or data schema definition.
// Schemas define the structure and constraints of data.
//
// Example:
//
//	schema := &Schema{
//	    Name:     "api_v1",
//	    Type:     "jsonschema",
//	    Version:  "1.0.0",
//	}
//
// Identifying Properties:
//   - name (required): Schema name
//
// Relationships:
//   - None (root node)
type Schema struct {
	// Name is the schema name.
	// This is an identifying property and is required.
	Name string

	// Type is the schema type.
	// Optional. Common values: "database", "jsonschema", "avro", "protobuf"
	Type string

	// Version is the schema version.
	// Optional.
	Version string

	// Definition is the schema definition content.
	// Optional.
	Definition string

	// Description is a description of the schema.
	// Optional.
	Description string
}

func (s *Schema) NodeType() string { return "schema" }

func (s *Schema) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: s.Name,
	}
}

func (s *Schema) Properties() map[string]any {
	props := s.IdentifyingProperties()
	if s.Type != "" {
		props["type"] = s.Type
	}
	if s.Version != "" {
		props["version"] = s.Version
	}
	if s.Definition != "" {
		props["definition"] = s.Definition
	}
	if s.Description != "" {
		props[graphrag.PropDescription] = s.Description
	}
	return props
}

func (s *Schema) ParentRef() *NodeRef      { return nil }
func (s *Schema) RelationshipType() string { return "" }

// DataPipeline represents a data processing pipeline.
// Pipelines orchestrate data transformation and movement.
//
// Example:
//
//	pipeline := &DataPipeline{
//	    Name:        "etl-daily",
//	    Type:        "airflow",
//	    Schedule:    "0 2 * * *",
//	    State:       "active",
//	}
//
// Identifying Properties:
//   - name (required): Pipeline name
//
// Relationships:
//   - None (root node)
type DataPipeline struct {
	// Name is the pipeline name.
	// This is an identifying property and is required.
	Name string

	// Type is the pipeline type or orchestrator.
	// Optional. Common values: "airflow", "dataflow", "glue", "databricks"
	Type string

	// Schedule is the execution schedule (cron format).
	// Optional. Example: "0 2 * * *"
	Schedule string

	// State is the current pipeline state.
	// Optional. Common values: "active", "paused", "failed"
	State string

	// LastRun is when the pipeline last executed.
	// Optional. Unix timestamp.
	LastRun int64

	// Description is a description of the pipeline.
	// Optional.
	Description string
}

func (d *DataPipeline) NodeType() string { return "data_pipeline" }

func (d *DataPipeline) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: d.Name,
	}
}

func (d *DataPipeline) Properties() map[string]any {
	props := d.IdentifyingProperties()
	if d.Type != "" {
		props["type"] = d.Type
	}
	if d.Schedule != "" {
		props["schedule"] = d.Schedule
	}
	if d.State != "" {
		props[graphrag.PropState] = d.State
	}
	if d.LastRun > 0 {
		props["last_run"] = d.LastRun
	}
	if d.Description != "" {
		props[graphrag.PropDescription] = d.Description
	}
	return props
}

func (d *DataPipeline) ParentRef() *NodeRef      { return nil }
func (d *DataPipeline) RelationshipType() string { return "" }
