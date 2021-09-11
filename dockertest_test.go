package dockertest

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/ory/dockertest/v3"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("postgres", "12", []string{"POSTGRES_PASSWORD=postgres"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error

		var connectString = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", "localhost", resource.GetPort("5432/tcp"), "postgres", "postgres", "postgres")
		fmt.Println("dsn:", connectString)
		db, err = sql.Open("postgres", connectString)
		if err != nil {
			return err
		}
		err = db.Ping()
		fmt.Println("ping err:", err)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestSomething(t *testing.T) {
	rows, err := db.Query("SELECT schemaname, tablename FROM pg_catalog.pg_tables;")
	if err != nil {
		t.Fatal(err)
	}
	var schemaname, tablename string
	for rows.Next() {
		err := rows.Scan(&schemaname, &tablename)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("schemaname: %s, tablename: %s\n", schemaname, tablename)
	}
}
