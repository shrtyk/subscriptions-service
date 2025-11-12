package postgres

import (
	"testing"

	"github.com/shrtyk/subscriptions-service/internal/config"
)

func TestBuildDSN(t *testing.T) {
	testCases := []struct {
		name     string
		cfg      *config.PostgresCfg
		expected string
	}{
		{
			name: "Standard configuration",
			cfg: &config.PostgresCfg{
				User:     "testuser",
				Password: "testpassword",
				Host:     "localhost",
				Port:     "5432",
				DBName:   "test-db",
				SSLMode:  "disable",
			},
			expected: "postgres://testuser:testpassword@localhost:5432/test-db?sslmode=disable",
		},
		{
			name: "SSLMode require",
			cfg: &config.PostgresCfg{
				User:     "produser",
				Password: "prodpassword",
				Host:     "db.example.com",
				Port:     "5432",
				DBName:   "prod-db",
				SSLMode:  "require",
			},
			expected: "postgres://produser:prodpassword@db.example.com:5432/prod-db?sslmode=require",
		},
		{
			name: "User and password with special characters",
			cfg: &config.PostgresCfg{
				User:     "user@name",
				Password: "password#123",
				Host:     "127.0.0.1",
				Port:     "5433",
				DBName:   "app-db",
				SSLMode:  "allow",
			},
			expected: "postgres://user%40name:password%23123@127.0.0.1:5433/app-db?sslmode=allow",
		},
		{
			name: "Empty password",
			cfg: &config.PostgresCfg{
				User:     "nopassuser",
				Password: "",
				Host:     "localhost",
				Port:     "5432",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			expected: "postgres://nopassuser:@localhost:5432/testdb?sslmode=disable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dsn := buildDSN(tc.cfg)
			if dsn != tc.expected {
				t.Errorf("Expected DSN: %s, but got: %s", tc.expected, dsn)
			}
		})
	}
}
