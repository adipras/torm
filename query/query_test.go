package query_test

import (
	"log"
	"os"
	"testing"

	"github.com/adipras/torm"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

var testDB *torm.Torm

func TestMain(m *testing.M) {
	dsn := "root:bismillah@tcp(127.0.0.1:3306)/app_test"

	db, err := torm.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to test DB: %v", err)
	}

	testDB = db
	code := m.Run()
	db.Close()
	os.Exit(code)
}

func setupTable(t *testing.T) {
	_, err := testDB.DB.SQL.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(255),
		age INT
	)`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	_, err = testDB.DB.SQL.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("failed to clear table: %v", err)
	}
}

func TestCreateAndFindAndFirst(t *testing.T) {
	setupTable(t)

	// Test Create
	user := User{Name: "Totti", Age: 40}
	err := testDB.Create(&User{}, &user)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("expected user ID to be set after Create")
	}

	// Test Find (executor)
	var users []User
	err = testDB.Find(&User{}, &users)
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}

	// Test Find (chain)
	users = []User{}
	err = testDB.Model(&User{}).Where("age >= ?", 18).Find(&users)
	if err != nil {
		t.Fatalf("chain Find() failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user with age >= 18, got %d", len(users))
	}

	// Test First (executor)
	var found User
	err = testDB.First(&User{}, &found, "WHERE id = ?", user.ID)
	if err != nil {
		t.Fatalf("First() failed: %v", err)
	}
	if found.Name != user.Name {
		t.Errorf("expected name %s, got %s", user.Name, found.Name)
	}

	// Test First (chain)
	var found2 User
	err = testDB.Model(&User{}).Where("id = ?", user.ID).First(&found2)
	if err != nil {
		t.Fatalf("chain First() failed: %v", err)
	}
	if found2.Name != user.Name {
		t.Errorf("expected name %s, got %s", user.Name, found2.Name)
	}
}

func TestUpdateAndDelete(t *testing.T) {
	setupTable(t)

	// Insert one user
	user := User{Name: "Dybala", Age: 30}
	err := testDB.Create(&User{}, &user)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Test Update
	err = testDB.Update(&User{}, map[string]any{"age": 31}, "WHERE id = ?", user.ID)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	var updated User
	err = testDB.First(&User{}, &updated, "WHERE id = ?", user.ID)
	if err != nil {
		t.Fatalf("verify update First() failed: %v", err)
	}
	if updated.Age != 31 {
		t.Errorf("expected age 31, got %d", updated.Age)
	}

	// Test Delete
	err = testDB.Delete(&User{}, "WHERE id = ?", user.ID)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	var deleted User
	err = testDB.First(&User{}, &deleted, "WHERE id = ?", user.ID)
	if err == nil {
		t.Fatal("expected no rows after delete, but found one")
	} else if err != torm.ErrNoRows {
		t.Fatalf("expected ErrNoRows, got: %v", err)
	}
}

func TestRawSQL(t *testing.T) {
	setupTable(t)

	// Insert some data
	_, err := testDB.DB.SQL.Exec(`INSERT INTO users (name, age) VALUES 
		('Alice', 25), ('Bob', 17), ('Charlie', 20)`)
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	rows, err := testDB.RawSQL("SELECT * FROM users WHERE age >= ?", 18)
	if err != nil {
		t.Fatalf("RawSQL failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Age)
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 rows from RawSQL, got %d", count)
	}
}
