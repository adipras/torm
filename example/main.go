package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adipras/torm"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func main() {
	db, err := torm.Open("mysql", "root:bismillah@tcp(127.0.0.1:3306)/app_test")
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// TEST: Create()
	user := User{Name: "Dybala", Age: 30}
	err = db.Create(&User{}, &user)
	if err != nil {
		log.Fatalf("Create failed: %v", err)
	}
	if user.ID == 0 {
		log.Fatalf("expected user ID to be set after Create")
	}
	fmt.Println("✅ Create success.")

	// TEST: Find() with chain query builder
	var users []User
	err = db.Model(&User{}).
		Where("age >= ?", 18).
		Find(&users)
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}
	for _, u := range users {
		fmt.Printf("User: ID=%d, Name=%s, Age=%d\n", u.ID, u.Name, u.Age)
	}
	fmt.Println("✅ Query builder Find() success.")

	// TEST: Find() with executor
	users = []User{}
	err = db.Find(&User{}, &users)
	if err != nil {
		log.Fatalf("Find failed: %v", err)
	}
	for _, u := range users {
		fmt.Printf("User: ID=%d, Name=%s, Age=%d\n", u.ID, u.Name, u.Age)
	}
	fmt.Println("✅ Direct executor Find() success.")

	// TEST: First()
	var single User
	err = db.Model(&User{}).Where("name = ?", "Dybala").First(&single)
	if err != nil {
		log.Fatalf("First failed: %v", err)
	}
	fmt.Println("✅ Query builder First() success.")
	fmt.Printf("✅ First success: ID=%d, Name=%s, Age=%d\n", single.ID, single.Name, single.Age)

	// TEST: First() with executor
	var firstUser User
	err = db.First(&User{}, &firstUser, "WHERE age >= ?", 18)
	if err != nil {
		log.Fatalf("First failed: %v", err)
	}
	fmt.Println("✅ Direct executor First() success.")
	fmt.Printf("First user: %+v\n", firstUser)

	// TEST: Update()
	err = db.Update(&User{}, map[string]any{"age": 31}, "WHERE id = ?", user.ID)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}
	fmt.Println("✅ Update success.")
	// Verify update
	var updatedUser User
	err = db.First(&User{}, &updatedUser, "WHERE id = ?", user.ID)
	if err != nil {
		log.Fatalf("Failed to verify update: %v", err)
	}
	if updatedUser.Age != 31 {
		log.Fatalf("Expected age to be 31, got %d", updatedUser.Age)
	}
	fmt.Printf("Updated user: ID=%d, Name=%s, Age=%d\n", updatedUser.ID, updatedUser.Name, updatedUser.Age)

	// TEST: Delete()
	err = db.Delete(&User{}, "WHERE id = ?", user.ID)
	if err != nil {
		log.Fatalf("Delete failed: %v", err)
	}
	fmt.Println("✅ Delete success.")
	// Verify delete
	var deletedUser User
	err = db.First(&User{}, &deletedUser, "WHERE id = ?", user.ID)
	if err == nil {
		log.Fatalf("Expected no rows after delete, but found: %+v", deletedUser)
	} else if err != torm.ErrNoRows {
		log.Fatalf("Expected ErrNoRows after delete, got: %v", err)
	}
	// Verify that the user was deleted
	fmt.Println("✅ User successfully deleted.")

	// TEST: RawSQL (default context)
	rows, err := db.RawSQL("SELECT * FROM users WHERE age > ?", 20)
	if err != nil {
		log.Fatalf("RawSQL failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Age); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("RawSQL User: ID=%d, Name=%s, Age=%d\n", u.ID, u.Name, u.Age)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("RawSQL rows error: %v", err)
	}

	// TEST: RawSQLContext (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows2, err := db.RawSQLContext(ctx, "SELECT * FROM users WHERE age > ?", 20)
	if err != nil {
		log.Fatalf("RawSQLContext failed: %v", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var u User
		if err := rows2.Scan(&u.ID, &u.Name, &u.Age); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("RawSQLContext User: ID=%d, Name=%s, Age=%d\n", u.ID, u.Name, u.Age)
	}
	if err := rows2.Err(); err != nil {
		log.Fatalf("RawSQLContext rows error: %v", err)
	}

	fmt.Println("✅ RawSQL and RawSQLContext success.")

	// Verify that all queries executed successfully
	fmt.Println("🎉 All queries executed successfully.")
}
