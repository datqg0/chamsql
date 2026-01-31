package runner

import (
	"context"
	"database/sql"
	//"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"backend/configs"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
)

// DBType represents supported database types
type DBType string

const (
	DBTypePostgreSQL DBType = "postgresql"
	DBTypeMySQL      DBType = "mysql"
	DBTypeSQLServer  DBType = "sqlserver"
)

var (
	ErrUnsupportedDB    = errors.New("unsupported database type")
	ErrQueryTimeout     = errors.New("query execution timeout")
	ErrInvalidStatement = errors.New("only SELECT statements are allowed")
	ErrConnectionFailed = errors.New("failed to connect to sandbox database")
)

// QueryResult holds the result of a query execution
type QueryResult struct {
	Columns     []string        `json:"columns"`
	Rows        [][]interface{} `json:"rows"`
	RowCount    int             `json:"rowCount"`
	ExecutionMs int64           `json:"executionMs"`
	Error       string          `json:"error,omitempty"`
	ErrorType   string          `json:"errorType,omitempty"` // timeout, syntax, runtime
}

// CompareResult holds comparison between expected and actual results
type CompareResult struct {
	IsCorrect     bool   `json:"isCorrect"`
	Message       string `json:"message,omitempty"`
	ExpectedRows  int    `json:"expectedRows"`
	ActualRows    int    `json:"actualRows"`
	MismatchIndex int    `json:"mismatchIndex,omitempty"` // First row with mismatch (-1 if none)
}

// Runner interface for query execution
type Runner interface {
	Execute(ctx context.Context, dbType DBType, query string) (*QueryResult, error)
	ExecuteWithSetup(ctx context.Context, dbType DBType, setupSQL, query string) (*QueryResult, error)
	Compare(expected, actual *QueryResult, orderMatters bool) *CompareResult
}

// runner implements Runner
type runner struct {
	cfg         *configs.Config
	connections map[DBType]*sql.DB
}

// NewRunner creates a new query runner
func NewRunner(cfg *configs.Config) (Runner, error) {
	r := &runner{
		cfg:         cfg,
		connections: make(map[DBType]*sql.DB),
	}

	// Initialize connections lazily on first use
	return r, nil
}

// getConnection returns a connection to the specified database type
func (r *runner) getConnection(dbType DBType) (*sql.DB, error) {
	if conn, exists := r.connections[dbType]; exists {
		if err := conn.Ping(); err == nil {
			return conn, nil
		}
		// Connection is stale, close and reconnect
		conn.Close()
		delete(r.connections, dbType)
	}

	var dsn string
	var driver string

	switch dbType {
	case DBTypePostgreSQL:
		dsn = r.cfg.SandboxPostgresURI
		driver = "postgres"
	case DBTypeMySQL:
		dsn = r.cfg.SandboxMySQLURI
		driver = "mysql"
	case DBTypeSQLServer:
		dsn = r.cfg.SandboxSQLServerURI
		driver = "sqlserver"
	default:
		return nil, ErrUnsupportedDB
	}

	if dsn == "" {
		return nil, fmt.Errorf("%w: %s sandbox not configured", ErrConnectionFailed, dbType)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	r.connections[dbType] = db
	return db, nil
}

// ValidateQuery checks if the query is a SELECT statement only
func ValidateQuery(query string) error {
	trimmed := strings.TrimSpace(strings.ToUpper(query))

	// Block dangerous statements
	forbidden := []string{"INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE", "GRANT", "REVOKE"}
	for _, stmt := range forbidden {
		if strings.HasPrefix(trimmed, stmt) {
			return ErrInvalidStatement
		}
	}

	// Must start with SELECT or WITH (for CTEs)
	if !strings.HasPrefix(trimmed, "SELECT") && !strings.HasPrefix(trimmed, "WITH") {
		return ErrInvalidStatement
	}

	return nil
}

// queryer allows running queries on *sql.DB or *sql.Tx
type queryer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// Execute runs a query on the specified database
func (r *runner) Execute(ctx context.Context, dbType DBType, query string) (*QueryResult, error) {
	// Validate query
	if err := ValidateQuery(query); err != nil {
		return &QueryResult{
			Error:     err.Error(),
			ErrorType: "validation",
		}, err
	}

	db, err := r.getConnection(dbType)
	if err != nil {
		return &QueryResult{
			Error:     err.Error(),
			ErrorType: "connection",
		}, err
	}

	return r.executeInternal(ctx, db, query)
}

func (r *runner) executeInternal(ctx context.Context, q queryer, query string) (*QueryResult, error) {
	// Create context with timeout
	timeout := time.Duration(r.cfg.QueryTimeoutSeconds) * time.Second
	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	startTime := time.Now()

	rows, err := q.QueryContext(queryCtx, query)
	if err != nil {
		errorType := "runtime"
		if errors.Is(err, context.DeadlineExceeded) {
			errorType = "timeout"
			err = ErrQueryTimeout
		}
		return &QueryResult{
			ExecutionMs: time.Since(startTime).Milliseconds(),
			Error:       err.Error(),
			ErrorType:   errorType,
		}, err
	}
	defer rows.Close()

	result, err := r.scanRows(rows)
	if err != nil {
		return &QueryResult{
			ExecutionMs: time.Since(startTime).Milliseconds(),
			Error:       err.Error(),
			ErrorType:   "runtime",
		}, err
	}

	result.ExecutionMs = time.Since(startTime).Milliseconds()

	// Limit rows
	if len(result.Rows) > r.cfg.QueryMaxRows {
		result.Rows = result.Rows[:r.cfg.QueryMaxRows]
		result.RowCount = r.cfg.QueryMaxRows
	}

	return result, nil
}

// ExecuteWithSetup runs setup SQL before the query (for problems with init_script)
func (r *runner) ExecuteWithSetup(ctx context.Context, dbType DBType, setupSQL, query string) (*QueryResult, error) {
	// Validate query first (we don't validate setupSQL as it contains DDL)
	if err := ValidateQuery(query); err != nil {
		return &QueryResult{
			Error:     err.Error(),
			ErrorType: "validation",
		}, err
	}

	db, err := r.getConnection(dbType)
	if err != nil {
		return &QueryResult{
			Error:     err.Error(),
			ErrorType: "connection",
		}, err
	}

	// Run setup in a transaction that will be rolled back
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return &QueryResult{
			Error:     err.Error(),
			ErrorType: "connection",
		}, err
	}
	defer tx.Rollback() // Always rollback to keep sandbox clean

	// Execute setup SQL - Split by semicolon to handle multiple statements
	if setupSQL != "" {
		stmts := strings.Split(setupSQL, ";")
		for _, stmt := range stmts {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				return &QueryResult{
					Error:     fmt.Sprintf("setup error: %v (stmt: %s)", err, stmt),
					ErrorType: "setup",
				}, err
			}
		}
	}

	// Execute the actual query within the same transaction
	return r.executeInternal(ctx, tx, query)
}

// scanRows converts sql.Rows to QueryResult (All values as Strings)
func (r *runner) scanRows(rows *sql.Rows) (*QueryResult, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := &QueryResult{
		Columns: columns,
		Rows:    make([][]interface{}, 0),
	}

	// Use RawBytes to get everything as bytes/string
	raw := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(raw))
	for i := range raw {
		scanArgs[i] = &raw[i]
	}

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		// Convert to string immediately
		row := make([]interface{}, len(columns))
		for i, col := range raw {
			if col == nil {
				row[i] = "" // Handle NULL as empty string? Or "NULL"? Reference uses string(col) which might be "" for nil slice? 
				// sql.RawBytes is a slice. If it's nil/empty, string(col) is "".
				// Let's assume empty string for simplicity matching reference.
			} else {
				row[i] = string(col)
			}
		}
		result.Rows = append(result.Rows, row)
	}

	result.RowCount = len(result.Rows)
	return result, rows.Err()
}

// convertValue converts database values to JSON-friendly types
func convertValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339)
	default:
		return val
	}
}



// Compare compares expected and actual query results using "String & Sorted" logic
func (r *runner) Compare(expected, actual *QueryResult, orderMatters bool) *CompareResult {
	result := &CompareResult{
		ExpectedRows:  expected.RowCount,
		ActualRows:    actual.RowCount,
		MismatchIndex: -1,
	}

	// Check for errors
	if actual.Error != "" {
		result.IsCorrect = false
		result.Message = fmt.Sprintf("Query error: %s", actual.Error)
		return result
	}

	// NOTE: We IGNORE column names and count to strict match sql-judge
	// But different column counts usually imply wrong query. 
	// sql-judge checks DeepEqual of [][]string, which implies inner slice length must match.
	// So we implicitly check column count via row comparison.

	// Compare row count
	if expected.RowCount != actual.RowCount {
		result.IsCorrect = false
		result.Message = fmt.Sprintf("Row count mismatch: expected %d, got %d",
			expected.RowCount, actual.RowCount)
		return result
	}

	// Clone rows for sorting to avoid mutating original result
	expRows := cloneRows(expected.Rows)
	actRows := cloneRows(actual.Rows)

	// ALWAYS Sort rows by string representation (ignoring orderMatters flag)
	sortRows(expRows)
	sortRows(actRows)

	// Compare sorted rows
	if !reflect.DeepEqual(expRows, actRows) {
		result.IsCorrect = false
		result.Message = "Result mismatch (values do not match)"
		
		// Find first mismatch for detail (optional, simplified)
		for i := range expRows {
			if i >= len(actRows) || !reflect.DeepEqual(expRows[i], actRows[i]) {
				result.MismatchIndex = i
				break
			}
		}
		return result
	}

	result.IsCorrect = true
	result.Message = "Correct!"
	return result
}

func cloneRows(rows [][]interface{}) [][]interface{} {
	newRows := make([][]interface{}, len(rows))
	for i, r := range rows {
		newRow := make([]interface{}, len(r))
		copy(newRow, r)
		newRows[i] = newRow
	}
	return newRows
}

func sortRows(rows [][]interface{}) {
	sort.Slice(rows, func(i, j int) bool {
		// Simple string representation sort
		return fmt.Sprint(rows[i]) < fmt.Sprint(rows[j])
	})
}

// Close closes all database connections
func (r *runner) Close() error {
	for _, db := range r.connections {
		db.Close()
	}
	r.connections = make(map[DBType]*sql.DB)
	return nil
}
