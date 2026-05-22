package main

import (
	"fmt"
	stdlog "log"
	"time"

	"github.com/dracory/neat"
)

// This example demonstrates different configuration options for the database connection
func main() {
	if err := RunExample(); err != nil {
		stdlog.Fatalf("Example failed: %v", err)
	}
}

// RunExample demonstrates different configuration options
func RunExample() error {
	// Example 1: Using DSN string (simplest approach)
	fmt.Println("=== Configuration with DSN ===")
	db1, err := neat.NewFromDSN("sqlite://./example.db")
	if err != nil {
		return fmt.Errorf("error with DSN config: %w", err)
	}
	defer db1.Close()
	fmt.Println("Database connected via DSN")

	// Example 2: Using configuration struct with connection pooling
	fmt.Println("\n=== Configuration with Pool Settings ===")
	cfg := neat.DBConfig{
		Default: "sqlite",
		Connections: map[string]neat.ConnectionConfig{
			"sqlite": {
				Driver:   "sqlite",
				Database: "./example.db",
			},
		},
		Pool: neat.PoolConfig{
			MaxIdleConns:    10,
			MaxOpenConns:    100,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: time.Hour,
		},
	}

	db2, err := neat.New(cfg)
	if err != nil {
		return fmt.Errorf("error with pool config: %w", err)
	}
	defer db2.Close()
	fmt.Println("Database connected with pool configuration")

	// Example 3: Multiple database connections
	fmt.Println("\n=== Multiple Database Connections ===")
	multiCfg := neat.DBConfig{
		Default: "sqlite",
		Connections: map[string]neat.ConnectionConfig{
			"sqlite": {
				Driver:   "sqlite",
				Database: "./local.db",
			},
			"postgres": {
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "primary_db",
				Username: "user",
				Password: "pass",
				SSLMode:  "disable",
			},
			"mysql": {
				Driver:   "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "secondary_db",
				Username: "user",
				Password: "pass",
				Charset:  "utf8mb4",
			},
		},
	}

	db4, err := neat.New(multiCfg)
	if err != nil {
		return fmt.Errorf("error with multi-database config: %w", err)
	}
	defer db4.Close()
	fmt.Println("Multiple databases configured successfully")

	// Example 4: MySQL-specific configuration
	fmt.Println("\n=== MySQL-Specific Configuration ===")
	mysqlCfg := neat.DBConfig{
		Default: "mysql",
		Connections: map[string]neat.ConnectionConfig{
			"mysql": {
				Driver:   "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Username: "root",
				Password: "password",
				Charset:  "utf8mb4",
			},
		},
	}

	db5, err := neat.New(mysqlCfg)
	if err != nil {
		return fmt.Errorf("error with MySQL config: %w", err)
	}
	defer db5.Close()
	fmt.Println("MySQL configured successfully")

	// Example 5: SQL Server configuration
	fmt.Println("\n=== SQL Server Configuration ===")
	sqlserverCfg := neat.DBConfig{
		Default: "sqlserver",
		Connections: map[string]neat.ConnectionConfig{
			"sqlserver": {
				Driver:   "sqlserver",
				Host:     "localhost",
				Port:     1433,
				Database: "mydb",
				Username: "sa",
				Password: "password",
			},
		},
	}

	db6, err := neat.New(sqlserverCfg)
	if err != nil {
		return fmt.Errorf("error with SQL Server config: %w", err)
	}
	defer db6.Close()
	fmt.Println("SQL Server configured successfully")

	// Example 6: Debug mode
	fmt.Println("\n=== Debug Mode Configuration ===")
	debugCfg := neat.DBConfig{
		Default: "sqlite",
		Connections: map[string]neat.ConnectionConfig{
			"sqlite": {
				Driver:   "sqlite",
				Database: "./example.db",
			},
		},
		Debug: true,
	}

	db7, err := neat.New(debugCfg)
	if err != nil {
		return fmt.Errorf("error with debug config: %w", err)
	}
	defer db7.Close()
	fmt.Println("Database configured in debug mode")

	return nil
}
