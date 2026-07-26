package action_test

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"github.com/anteiro255/gedis/internal/action"
	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/pkg/protocol"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

func setupTestDB() (*db.DB, db.Key) {
	database := db.NewDB()
	key := db.Key([16]byte{1, 2, 3})
	return database, key
}

func TestAction_Perform_SetAndGet(t *testing.T) {
	database, key := setupTestDB()
	val := []byte("value123")

	// Set Action
	setAction := action.Action{
		DB:         database,
		ActionType: protocol.Set,
		Key:        key,
		Body:       val,
	}
	body, sts := setAction.Perform()
	if sts != status.OK {
		t.Fatalf("Set expected status.OK, got %v", sts)
	}
	if body != nil {
		t.Errorf("Set expected nil body, got %v", body)
	}

	// Get Action
	getAction := action.Action{
		DB:         database,
		ActionType: protocol.Get,
		Key:        key,
	}
	gotVal, sts := getAction.Perform()
	if sts != status.OK {
		t.Fatalf("Get expected status.OK, got %v", sts)
	}
	if !bytes.Equal(gotVal, val) {
		t.Errorf("Get expected %q, got %q", string(val), string(gotVal))
	}
}

func TestAction_Perform_DelAndExist(t *testing.T) {
	database, key := setupTestDB()
	database.Set(key, db.Val([]byte("data")))

	// Exist check
	existAction := action.Action{
		DB:         database,
		ActionType: protocol.Exist,
		Key:        key,
	}
	_, sts := existAction.Perform()
	if sts != status.OK {
		t.Errorf("Exist expected status.OK, got %v", sts)
	}

	// Del Action
	delAction := action.Action{
		DB:         database,
		ActionType: protocol.Del,
		Key:        key,
	}
	_, sts = delAction.Perform()
	if sts != status.OK {
		t.Errorf("Del expected status.OK, got %v", sts)
	}

	// Exist check after Del
	_, sts = existAction.Perform()
	if sts != status.NoSuchKey {
		t.Errorf("Exist after Del expected status.NoSuchKey, got %v", sts)
	}
}

func TestAction_Perform_TTL(t *testing.T) {
	database, key := setupTestDB()
	database.Set(key, db.Val([]byte("ttl_data")))

	// TTL_Set invalid body size
	ttlSetBad := action.Action{
		DB:         database,
		ActionType: protocol.TTL_Set,
		Key:        key,
		Body:       []byte{1, 2}, // invalid size != 4
	}
	_, sts := ttlSetBad.Perform()
	if sts != status.WrongInput {
		t.Errorf("TTL_Set with invalid body size expected status.WrongInput, got %v", sts)
	}

	// TTL_Set valid body
	ttlBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ttlBytes, 10)
	ttlSetValid := action.Action{
		DB:         database,
		ActionType: protocol.TTL_Set,
		Key:        key,
		Body:       ttlBytes,
	}
	_, sts = ttlSetValid.Perform()
	if sts != status.OK {
		t.Fatalf("TTL_Set expected status.OK, got %v", sts)
	}

	// TTL_Get
	ttlGet := action.Action{
		DB:         database,
		ActionType: protocol.TTL_Get,
		Key:        key,
	}
	resBody, sts := ttlGet.Perform()
	if sts != status.OK {
		t.Fatalf("TTL_Get expected status.OK, got %v", sts)
	}
	if len(resBody) != 4 {
		t.Fatalf("TTL_Get expected 4 bytes response, got %d", len(resBody))
	}
	remainingTTL := binary.BigEndian.Uint32(resBody)
	if remainingTTL == 0 || remainingTTL > 10 {
		t.Errorf("TTL_Get expected ~10, got %d", remainingTTL)
	}

	// TTL_Del
	ttlDel := action.Action{
		DB:         database,
		ActionType: protocol.TTL_Del,
		Key:        key,
	}
	_, sts = ttlDel.Perform()
	if sts != status.OK {
		t.Errorf("TTL_Del expected status.OK, got %v", sts)
	}

	// TTL_Get after TTL_Del
	_, sts = ttlGet.Perform()
	if sts != status.NoTTL {
		t.Errorf("TTL_Get after TTL_Del expected status.NoTTL, got %v", sts)
	}
}

func TestAction_Perform_UnknownAction(t *testing.T) {
	database, key := setupTestDB()
	unknownAction := action.Action{
		DB:         database,
		ActionType: protocol.ActionType(255),
		Key:        key,
	}
	_, sts := unknownAction.Perform()
	if sts != status.WrongInput {
		t.Errorf("Unknown action expected status.WrongInput, got %v", sts)
	}
}

// Quiet unused import check
var _ = time.Millisecond
