package helpers

// SampleContent provides test data fixtures
type SampleContent struct {
	Name        string
	Content     string
	Description string
}

// GetCodeSamples returns sample code for testing
func GetCodeSamples() []SampleContent {
	return []SampleContent{
		{
			Name:        "go_simple",
			Description: "Simple Go code",
			Content: `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`,
		},
		{
			Name:        "go_complex",
			Description: "Complex Go code with imports and functions",
			Content: `package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// User represents a user in the system
type User struct {
	ID        int64     ` + "`json:\"id\"`" + `
	Name      string    ` + "`json:\"name\"`" + `
	Email     string    ` + "`json:\"email\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
}

// UserService handles user operations
type UserService struct {
	db *Database
}

// NewUserService creates a new user service
func NewUserService(db *Database) *UserService {
	return &UserService{db: db}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id int64) (*User, error) {
	// Implementation here
	return nil, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, user *User) error {
	// Validate input
	if user.Name == "" {
		return fmt.Errorf("name is required")
	}
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}
	
	// Insert into database
	return s.db.Insert(ctx, user)
}

func main() {
	log.Println("Starting application...")
	
	db := &Database{}
	service := NewUserService(db)
	
	ctx := context.Background()
	user := &User{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	
	if err := service.CreateUser(ctx, user); err != nil {
		log.Fatal(err)
	}
	
	log.Println("User created successfully")
}
`,
		},
		{
			Name:        "json_data",
			Description: "Sample JSON data",
			Content: `{
  "users": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "active": true
    },
    {
      "id": 2,
      "name": "Jane Smith",
      "email": "jane@example.com",
      "active": false
    }
  ],
  "total": 2,
  "page": 1,
  "per_page": 10
}
`,
		},
		{
			Name:        "log_output",
			Description: "Sample log output",
			Content: `2024-01-15T10:30:00Z INFO Starting application
2024-01-15T10:30:01Z INFO Connected to database
2024-01-15T10:30:02Z WARN High memory usage: 85%
2024-01-15T10:30:03Z INFO Processing request #1234
2024-01-15T10:30:04Z ERROR Failed to process request: timeout
2024-01-15T10:30:05Z INFO Retrying request #1234
2024-01-15T10:30:06Z INFO Request #1234 completed successfully
2024-01-15T10:30:07Z INFO Shutting down gracefully
`,
		},
	}
}

// GetLargeContent returns large content for stress testing
func GetLargeContent() string {
	content := ""
	for i := 0; i < 1000; i++ {
		content += "This is line number " + string(rune('0'+i%10)) + " with some additional content to make it longer.\n"
	}
	return content
}

// GetRepetitiveContent returns content with repetition
func GetRepetitiveContent() string {
	return `ERROR: Connection failed
ERROR: Connection failed
ERROR: Connection failed
ERROR: Connection failed
ERROR: Connection failed
WARNING: Retrying...
WARNING: Retrying...
WARNING: Retrying...
SUCCESS: Connected
`
}
