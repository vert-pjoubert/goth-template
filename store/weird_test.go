package store

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

// Test structs
type TestEvent struct {
	ID    int64       `db:"id"`
	Name  string      `db:"name"`
	Roles []string    `db:"roles"`
	Meta  interface{} `db:"meta"`
}

type TestRole struct {
	ID          int64    `db:"id"`
	Name        string   `db:"name"`
	Permissions []string `db:"permissions"`
}

type TestServer struct {
	ID    int64             `db:"id"`
	Name  string            `db:"name"`
	Roles []string          `db:"roles"`
	Meta  map[string]string `db:"meta"`
}

// Generic utility functions
func loadIntoMap(db *sqlx.DB, tableName string, fieldName string, fieldValue interface{}) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", tableName, fieldName)
	rows, err := db.Queryx(query, fieldValue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no rows found")
	}

	result := make(map[string]interface{})
	if err := rows.MapScan(result); err != nil {
		return nil, err
	}

	return result, nil
}

func handleJSONFields(m map[string]interface{}, jsonFields []string) error {
	for _, field := range jsonFields {
		if value, ok := m[field]; ok {
			switch v := value.(type) {
			case []byte:
				if len(v) == 0 {
					m[field] = nil
					continue
				}
				var temp interface{}
				if err := json.Unmarshal(v, &temp); err != nil {
					return err
				}
				if temp == nil {
					m[field] = nil
				} else {
					m[field] = temp
				}
			case string:
				if len(v) == 0 {
					m[field] = nil
					continue
				}
				var temp interface{}
				if err := json.Unmarshal([]byte(v), &temp); err != nil {
					return err
				}
				if temp == nil {
					m[field] = nil
				} else {
					m[field] = temp
				}
			}
		} else {
			m[field] = nil
		}
	}
	return nil
}

func mapToStruct(m map[string]interface{}, dest interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func fetchAndConvert(db *sqlx.DB, tableName, fieldName string, fieldValue interface{}, dest interface{}, jsonFields []string) error {
	m, err := loadIntoMap(db, tableName, fieldName, fieldValue)
	if err != nil {
		return err
	}

	if err := handleJSONFields(m, jsonFields); err != nil {
		return err
	}

	if err := mapToStruct(m, dest); err != nil {
		return err
	}

	return nil
}

// Unit Tests
func TestGenericLogic(t *testing.T) {
	// Setup logging to a file
	logFile, err := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	logger.Println("Starting TestGenericLogic")

	// Setup in-memory SQLite database
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		logger.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create test tables
	schema := `
	CREATE TABLE events (
		id INTEGER PRIMARY KEY,
		name TEXT,
		roles TEXT,
		meta TEXT
	);
	CREATE TABLE roles (
		id INTEGER PRIMARY KEY,
		name TEXT,
		permissions TEXT
	);
	CREATE TABLE servers (
		id INTEGER PRIMARY KEY,
		name TEXT,
		roles TEXT,
		meta TEXT
	);
	`
	db.MustExec(schema)
	logger.Println("Created test tables")

	// Insert test data
	insertTestData(db, t, logger)

	// Test fetching and converting data
	t.Run("TestEvent", func(t *testing.T) {
		event := new(TestEvent)
		err := fetchAndConvert(db, "events", "id", 1, event, []string{"roles", "meta"})
		assert.NoError(t, err)
		expectedMeta := map[string]interface{}{"description": "Test event meta", "nested": map[string]interface{}{"key": "value"}}
		assert.Equal(t, &TestEvent{ID: 1, Name: "Test Event", Roles: []string{"admin", "user"}, Meta: expectedMeta}, event)
		logger.Println("TestEvent passed")
	})

	t.Run("TestRole", func(t *testing.T) {
		role := new(TestRole)
		err := fetchAndConvert(db, "roles", "id", 1, role, []string{"permissions"})
		assert.NoError(t, err)
		assert.Equal(t, &TestRole{ID: 1, Name: "Test Role", Permissions: []string{"read", "write"}}, role)
		logger.Println("TestRole passed")
	})

	t.Run("TestServer", func(t *testing.T) {
		server := new(TestServer)
		err := fetchAndConvert(db, "servers", "id", 1, server, []string{"roles", "meta"})
		assert.NoError(t, err)
		expectedMeta := map[string]string{"env": "production", "region": "us-west"}
		assert.Equal(t, &TestServer{ID: 1, Name: "Test Server", Roles: []string{"admin", "user"}, Meta: expectedMeta}, server)
		logger.Println("TestServer passed")
	})

	t.Run("TestInvalidJSON", func(t *testing.T) {
		event := new(TestEvent)
		err := fetchAndConvert(db, "events", "id", 2, event, []string{"roles", "meta"})
		assert.Error(t, err)
		logger.Println("TestInvalidJSON passed")
	})

	t.Run("TestMissingFields", func(t *testing.T) {
		event := new(TestEvent)
		err := fetchAndConvert(db, "events", "id", 3, event, []string{"roles", "meta"})
		assert.NoError(t, err)
		assert.Equal(t, &TestEvent{ID: 3, Name: "Incomplete Event", Roles: []string{}, Meta: nil}, event)
		logger.Println("TestMissingFields passed")
	})

	t.Run("TestLargeDataSet", func(t *testing.T) {
		// Insert a large number of events
		for i := 4; i <= 1003; i++ {
			_, err := db.Exec(`INSERT INTO events (name, roles, meta) VALUES (?, ?, ?)`, fmt.Sprintf("Event %d", i), `["user"]`, `{"description": "Large data set event"}`)
			if err != nil {
				t.Fatalf("Failed to insert large data set event: %v", err)
			}
		}
		event := new(TestEvent)
		err := fetchAndConvert(db, "events", "id", 1003, event, []string{"roles", "meta"})
		assert.NoError(t, err)
		expectedMeta := map[string]interface{}{"description": "Large data set event"}
		assert.Equal(t, &TestEvent{ID: 1003, Name: "Event 1003", Roles: []string{"user"}, Meta: expectedMeta}, event)
		logger.Println("TestLargeDataSet passed")
	})

	logger.Println("Finished TestGenericLogic")
}

func insertTestData(db *sqlx.DB, t *testing.T, logger *log.Logger) {
	rolesJSON := `["admin", "user"]`
	permissionsJSON := `["read", "write"]`
	metaJSON := `{"description": "Test event meta", "nested": {"key": "value"}}`
	invalidJSON := `{"description": "Invalid JSON, "nested": {"key": "value"}}`
	missingFieldsJSON := `{}`

	_, err := db.Exec(`INSERT INTO events (name, roles, meta) VALUES (?, ?, ?)`, "Test Event", rolesJSON, metaJSON)
	if err != nil {
		logger.Fatalf("Failed to insert event: %v", err)
	}

	_, err = db.Exec(`INSERT INTO roles (name, permissions) VALUES (?, ?)`, "Test Role", permissionsJSON)
	if err != nil {
		logger.Fatalf("Failed to insert role: %v", err)
	}

	_, err = db.Exec(`INSERT INTO servers (name, roles, meta) VALUES (?, ?, ?)`, "Test Server", rolesJSON, `{"env": "production", "region": "us-west"}`)
	if err != nil {
		logger.Fatalf("Failed to insert server: %v", err)
	}

	_, err = db.Exec(`INSERT INTO events (name, roles, meta) VALUES (?, ?, ?)`, "Invalid Event", rolesJSON, invalidJSON)
	if err != nil {
		logger.Fatalf("Failed to insert invalid event: %v", err)
	}

	_, err = db.Exec(`INSERT INTO events (name, roles, meta) VALUES (?, ?, ?)`, "Incomplete Event", `[]`, missingFieldsJSON)
	if err != nil {
		logger.Fatalf("Failed to insert incomplete event: %v", err)
	}

	logger.Println("Inserted test data")
}
