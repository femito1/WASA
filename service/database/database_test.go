package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// setupTestDB initializes an in-memory SQLite DB and returns an AppDatabase instance.
func setupTestDB(t *testing.T) AppDatabase {
	t.Helper()

	// Use an in-memory database for testing.
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite database: %v", err)
	}

	appdb, err := New(sqlDB)
	if err != nil {
		t.Fatalf("failed to initialize AppDatabase: %v", err)
	}
	return appdb
}

func TestUserOperations(t *testing.T) {
	db := setupTestDB(t)

	// Test CreateUser.
	user := User{Username: "testuser"}
	createdUser, err := db.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if createdUser.Id == 0 {
		t.Errorf("expected non-zero user id, got %d", createdUser.Id)
	}

	// Test GetUserId.
	fetchedUser, err := db.GetUserId("testuser")
	if err != nil {
		t.Fatalf("GetUserId failed: %v", err)
	}
	if fetchedUser.Id != createdUser.Id {
		t.Errorf("expected user id %d, got %d", createdUser.Id, fetchedUser.Id)
	}

	// Test SetUsername.
	updatedUser, err := db.SetUsername(createdUser, "newname")
	if err != nil {
		t.Fatalf("SetUsername failed: %v", err)
	}
	if updatedUser.Username != "newname" {
		t.Errorf("expected username 'newname', got %s", updatedUser.Username)
	}

	// Test ListUsers (no filter).
	users, err := db.ListUsers("")
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}

func TestConversationOperations(t *testing.T) {
	db := setupTestDB(t)

	// Create two users.
	user1, err := db.CreateUser(User{Username: "user1"})
	if err != nil {
		t.Fatalf("CreateUser user1 failed: %v", err)
	}
	user2, err := db.CreateUser(User{Username: "user2"})
	if err != nil {
		t.Fatalf("CreateUser user2 failed: %v", err)
	}

	// Create a conversation with user1 as creator and user2 as additional member.
	conv, err := db.CreateConversation(user1, "Test Group", []uint64{user2.Id})
	if err != nil {
		t.Fatalf("CreateConversation failed: %v", err)
	}
	if conv.Id == 0 {
		t.Errorf("expected conversation id to be set")
	}
	if len(conv.Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(conv.Members))
	}

	// Test SetConversationName.
	conv, err = db.SetConversationName(user1.Id, conv.Id, "New Group Name")
	if err != nil {
		t.Fatalf("SetConversationName failed: %v", err)
	}
	if conv.Name != "New Group Name" {
		t.Errorf("expected conversation name 'New Group Name', got %s", conv.Name)
	}

	// Test AddUserToConversation (add a third user).
	user3, err := db.CreateUser(User{Username: "user3"})
	if err != nil {
		t.Fatalf("CreateUser user3 failed: %v", err)
	}
	conv, err = db.AddUserToConversation(user1.Id, conv.Id, user3.Id)
	if err != nil {
		t.Fatalf("AddUserToConversation failed: %v", err)
	}
	if len(conv.Members) != 3 {
		t.Errorf("expected 3 members, got %d", len(conv.Members))
	}

	// Test RemoveUserFromConversation (remove user2).
	err = db.RemoveUserFromConversation(user2.Id, conv.Id)
	if err != nil {
		t.Fatalf("RemoveUserFromConversation failed: %v", err)
	}
	conv, err = db.GetConversation(user1.Id, conv.Id, nil)
	if err != nil {
		t.Fatalf("GetConversation failed: %v", err)
	}
	if len(conv.Members) != 2 {
		t.Errorf("expected 2 members after removal, got %d", len(conv.Members))
	}
}

func TestMessageAndCommentOperations(t *testing.T) {
	db := setupTestDB(t)

	// Create users and a conversation.
	sender, err := db.CreateUser(User{Username: "sender"})
	if err != nil {
		t.Fatalf("CreateUser sender failed: %v", err)
	}
	receiver, err := db.CreateUser(User{Username: "receiver"})
	if err != nil {
		t.Fatalf("CreateUser receiver failed: %v", err)
	}
	conv, err := db.CreateConversation(sender, "Chat", []uint64{receiver.Id})
	if err != nil {
		t.Fatalf("CreateConversation failed: %v", err)
	}

	// Test CreateMessage.
	msg, err := db.CreateMessage(sender, conv.Id, "Hello World", "string")
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}
	if msg.Id == 0 {
		t.Errorf("expected message id to be set")
	}

	// Test ForwardMessage: forward the message (for simplicity, forward it back into the same conversation).
	err = db.ForwardMessage(receiver, conv.Id, msg.Id, conv.Id)
	if err != nil {
		t.Fatalf("ForwardMessage failed: %v", err)
	}

	// Test CommentMessage.
	commentId, err := db.CommentMessage(receiver, conv.Id, msg.Id, "Nice message!")
	if err != nil {
		t.Fatalf("CommentMessage failed: %v", err)
	}
	if commentId == 0 {
		t.Errorf("expected comment id to be set")
	}

	// Test DeleteComment.
	err = db.DeleteComment(receiver, conv.Id, msg.Id, commentId)
	if err != nil {
		t.Fatalf("DeleteComment failed: %v", err)
	}

	// Test DeleteMessage.
	err = db.DeleteMessage(sender, conv.Id, msg.Id)
	if err != nil {
		t.Fatalf("DeleteMessage failed: %v", err)
	}
}
