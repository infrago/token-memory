package tokenmemory

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	. "github.com/infrago/base"
	"github.com/infrago/token"
	"github.com/tidwall/buntdb"
)

type memoryDriver struct {
	mutex sync.Mutex
	db    *buntdb.DB
}

func init() {
	token.RegisterDriver("memory", &memoryDriver{})
}

func (d *memoryDriver) Configure(Map) {}

func (d *memoryDriver) Open() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.db != nil {
		return nil
	}
	db, err := buntdb.Open(":memory:")
	if err != nil {
		return err
	}
	d.db = db
	return nil
}

func (d *memoryDriver) Close() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.db == nil {
		return nil
	}
	err := d.db.Close()
	d.db = nil
	return err
}

func (d *memoryDriver) SavePayload(tokenID string, payload Map, exp int64) error {
	tokenID = strings.TrimSpace(tokenID)
	if tokenID == "" {
		return nil
	}
	db, err := d.ensureDB()
	if err != nil {
		return err
	}
	bts, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(d.keyPayload(tokenID), string(bts), d.setOptions(exp))
		return err
	})
}

func (d *memoryDriver) LoadPayload(tokenID string) (Map, bool, error) {
	tokenID = strings.TrimSpace(tokenID)
	if tokenID == "" {
		return nil, false, nil
	}
	db, err := d.ensureDB()
	if err != nil {
		return nil, false, err
	}
	var raw string
	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(d.keyPayload(tokenID))
		if err != nil {
			return err
		}
		raw = val
		return nil
	})
	if err == buntdb.ErrNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	out := Map{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

func (d *memoryDriver) DeletePayload(tokenID string) error {
	tokenID = strings.TrimSpace(tokenID)
	if tokenID == "" {
		return nil
	}
	db, err := d.ensureDB()
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(d.keyPayload(tokenID))
		if err == buntdb.ErrNotFound {
			return nil
		}
		return err
	})
}

func (d *memoryDriver) RevokeToken(token string, exp int64) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	db, err := d.ensureDB()
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(d.keyRevokeToken(token), "1", d.setOptions(exp))
		return err
	})
}

func (d *memoryDriver) RevokeTokenID(tokenID string, exp int64) error {
	tokenID = strings.TrimSpace(tokenID)
	if tokenID == "" {
		return nil
	}
	db, err := d.ensureDB()
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(d.keyRevokeTokenID(tokenID), "1", d.setOptions(exp))
		return err
	})
}

func (d *memoryDriver) RevokedToken(token string) (bool, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return false, nil
	}
	db, err := d.ensureDB()
	if err != nil {
		return false, err
	}
	err = db.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get(d.keyRevokeToken(token))
		return err
	})
	if err == buntdb.ErrNotFound {
		return false, nil
	}
	return err == nil, err
}

func (d *memoryDriver) RevokedTokenID(tokenID string) (bool, error) {
	tokenID = strings.TrimSpace(tokenID)
	if tokenID == "" {
		return false, nil
	}
	db, err := d.ensureDB()
	if err != nil {
		return false, err
	}
	err = db.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get(d.keyRevokeTokenID(tokenID))
		return err
	})
	if err == buntdb.ErrNotFound {
		return false, nil
	}
	return err == nil, err
}

func (d *memoryDriver) ensureDB() (*buntdb.DB, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.db != nil {
		return d.db, nil
	}
	db, err := buntdb.Open(":memory:")
	if err != nil {
		return nil, err
	}
	d.db = db
	return d.db, nil
}

func (d *memoryDriver) setOptions(exp int64) *buntdb.SetOptions {
	if exp <= 0 {
		return nil
	}
	ttl := time.Until(time.Unix(exp, 0))
	if ttl <= 0 {
		ttl = time.Second
	}
	return &buntdb.SetOptions{Expires: true, TTL: ttl}
}

func (d *memoryDriver) keyPayload(tokenID string) string {
	return "payload:" + tokenID
}

func (d *memoryDriver) keyRevokeToken(token string) string {
	return "revoke:token:" + token
}

func (d *memoryDriver) keyRevokeTokenID(tokenID string) string {
	return "revoke:tokenid:" + tokenID
}
