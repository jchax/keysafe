package keyctl

import (
	"encoding/json"
	"runtime"
	"strconv"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

const keysafe = "deposit-box"

type Service struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestNewKeySafe(t *testing.T) {
	keyring, err := NewKeySafe(keysafe)
	if err != nil {
		t.Errorf("New keyring failed: %v\n", err)
	}
	if err := keyring.Unlink(); err != nil {
		t.Errorf("Unlink new keyring failed: %v\n", err)
	}
}

func TestPeristentKeyring(t *testing.T) {
	keyring, err := NewKeySafe(keysafe)
	if err != nil {
		t.Errorf("Create peristent keyring failed: %v\n", err)
	}
	if err := keyring.Set("hello", []byte("world")); err != nil {
		t.Errorf("Set(%q) failed: %v\n", "hello", err)
	}
	val1, err := keyring.Get("hello")
	if err != nil {
		t.Errorf("Get(%q) failed: %v\b", "hello", err)
	}
	if string(val1) != "world" {
		t.Errorf("Wanted %q, got %q\n", "world", val1)
	}

	// Re-fetch keyring to make sure we get the same one
	keyring, err = NewKeySafe(keysafe)
	if err != nil {
		t.Errorf("Re-fetch persistent keyring failed: %v\n", err)
	}
	val2, err := keyring.Get("hello")
	if err != nil {
		t.Errorf("Re-get \"hello\" failed: %v\n", err)
	}
	if string(val1) != string(val2) {
		t.Errorf("Re-get, wanted %q, got %q\n", val1, val2)
	}
	if err := keyring.Unlink(); err != nil {
		t.Errorf("Unlink keyring failed: %v\n", err)
	}
}

func TestMultiStoreKeyring(t *testing.T) {
	m := make(map[string]string)
	m["spare-key"] = "under the doormat"
	m["combination"] = strconv.Itoa(271828)
	m["hello world"] = "你好，世界"
	keyring, err := NewKeySafe(keysafe)
	if err != nil {
		t.Errorf("Create keyring failed: %v\n", err)
	}
	for key, value := range m {
		if err := keyring.Set(key, []byte(value)); err != nil {
			t.Errorf("Set(%q, %q) failed: %v\n", key, value, err)
		}
	}
	for key, value := range m {
		value1, err := keyring.Get(key)
		if err != nil {
			t.Errorf("Get(%q) failed: %v\n", key, err)
		}
		if value != string(value1) {
			t.Errorf("Get(%q), wanted %q, got %q\n", key, value, value1)
		}
	}
	_, err = keyring.Get("nothing")
	if err == nil || err != unix.ENOKEY {
		t.Errorf("Unexpecred success or wrong error: %v\n", err)
	}
	if err := keyring.Unlink(); err != nil {
		t.Errorf("Unlink keyring failed: %v\n", err)
	}
}

func TestTimeout(t *testing.T) {
	keyring, err := NewKeySafe(keysafe)
	if err != nil {
		t.Errorf("Create keyring failed: %v\n", err)
	}
	key := "temporary-key"
	value := "under the doormat"
	if err := keyring.Set(key, []byte(value)); err != nil {
		t.Errorf("Set(%q, %q) failed: %v\n", key, value, err)
	}
	if err := keyring.SetTimeout(key, time.Second); err != nil {
		t.Errorf("SetTimeout(%q, 45) failed: %v\n", key, err)
	}
	if _, err := keyring.Get(key); err != nil {
		t.Errorf("Active Get(%q) failed: %v\n", key, err)
	}
	time.Sleep(1100 * time.Millisecond)
	if _, err := keyring.Get(key); err != ErrKeyExpired {
		t.Errorf("Expired Get(%q) failed: %v\n", key, err)
	}
	if err := keyring.Unlink(); err != nil {
		t.Errorf("Unlink keyring failed: %v\n", err)
	}
}

func TestDelete(t *testing.T) {
	keyring, err := NewKeySafe(keysafe)
	if err != nil {
		t.Errorf("Create keyring failed: %v\n", err)
	}
	key := "temporary-key"
	value := "under the doormat"
	if err := keyring.Set(key, []byte(value)); err != nil {
		t.Errorf("Set(%q, %q) failed: %v\n", key, value, err)
	}
	if _, err := keyring.Get(key); err != nil {
		t.Errorf("Active Get(%q) failed: %v\n", key, err)
	}
	if err := keyring.Del(key); err != nil {
		t.Errorf("Delete(%q) failed: %v\n", key, err)
	}
	if _, err := keyring.Get(key); err != ErrNoKey {
		t.Errorf("Deleted Get(%q) failed: %v\n", key, err)
	}
	if err := keyring.Unlink(); err != nil {
		t.Errorf("Unlink keyring failed: %v\n", err)
	}
}

func TestAfterUnlink(t *testing.T) {
	keyring, err := NewKeySafe(keysafe)
	if err != nil {
		t.Errorf("Create keyring failed: %v\n", err)
	}
	key := "temporary-key"
	value := "under the doormat"
	if err := keyring.Set(key, []byte(value)); err != nil {
		t.Errorf("Set(%q, %q) failed: %v\n", key, value, err)
	}
	if _, err := keyring.Get(key); err != nil {
		t.Errorf("Active Get(%q) failed: %v\n", key, err)
	}
	if err := keyring.Unlink(); err != nil {
		t.Errorf("Unlink failed: %v\n", err)
	}
	if _, err := keyring.Get(key); err != ErrNoKey {
		t.Errorf("Unlink Get(%q) failed: %v\n", key, err)
	}
	if err := keyring.Unlink(); err != ErrNoKey {
		t.Errorf("Unlink unlink keyring failed: %v\n", err)
	}
}

// I'm not convinced this test works.  In principle, the two go
// routines are in separate OS threads which should stop it working on
// the old keysafe implementation but that would seem to depend on a
// race condition.  Nonetheless, this test and the actual failing
// progra do at least pass.
func TestCrossThread(t *testing.T) {
	ch := make(chan *KeySafe)
	key := "temporary-key"
	value := "under the doormat"
	go func() {
		runtime.LockOSThread()
		keyring, err := NewKeySafe(keysafe)
		if err != nil {
			t.Errorf("Create1 keyring  failed: %v\n", err)
		}
		if err := keyring.Set(key, []byte(value)); err != nil {
			t.Errorf("Set1(%q, %q) failed: %v\n", key, value, err)
		}
		if _, err := keyring.Get(key); err != nil {
			t.Errorf("Get1(%q) failed: %v\n", key, err)
		}
		ch <- keyring
		<-ch
	}()
	keyring := <-ch
	go func() {
		runtime.LockOSThread()
		val, err := keyring.Get(key)
		if err != nil {
			t.Errorf("Get2(%q) failed: %v\n", key, err)
		}
		if string(val) != value {
			t.Errorf("Get2: expected %q, got %q\n", value, string(val))
		}
		close(ch)
	}()
	<-ch
	if err := keyring.Unlink(); err != nil {
		t.Errorf("Unlink keyring failed: %v\n", err)
	}
}

func TestJsonKeyring(t *testing.T) {
	m := make(map[string]*Service)
	keys := []string{"example", "秘密", "google.com"}
	m[keys[0]] = &Service{
		Server:   "https://example.com/foo",
		Username: "FRED",
		Password: "SekritSauce"}
	m[keys[1]] = &Service{
		Server:   "mail;秘密.com",
		Username: "fred.smith@example.com",
		Password: "A huge secret, 秘密"}
	m[keys[2]] = &Service{
		Server:   "mail;google.com",
		Username: "fred.jones@gmail.com",
		Password: "A password"}
	keyring, err := NewKeySafe(keysafe)
	if err != nil {
		t.Errorf("Create keyring failed: %v\n", err)
	}
	for key, value := range m {
		if err := keyring.SetService(key, value); err != nil {
			t.Errorf("Cannot set %v: %v\n", key, err)
		}
	}
	names, err := keyring.List()
	if err != nil {
		t.Errorf("Cannot get name list: %v\n", err)
	}
	if len(names) != len(keys) {
		t.Errorf("Expected %d names, got %d\n", len(keys), len(names))
	}
	for _, k := range names {
		service, err := keyring.GetService(k)
		if err != nil {
			t.Errorf("Cannot get %v: %v\n", k, err)
		}
		if *m[k] != *service {
			t.Errorf("JSON, wanted %v,\n\tgot %v\n", m[k], service)
		}
	}
	if err := keyring.Unlink(); err != nil {
		t.Errorf("Unlink keyring failed: %v\n", err)
	}
}

// Get a Service from the keyring
func (k *KeySafe) GetService(name string) (*Service, error) {
	val, err := k.Get(name)
	if err != nil {
		return nil, err
	}
	var s Service
	err = json.Unmarshal(val, &s)
	return &s, err
}

//Set a Service in the keyring
func (k *KeySafe) SetService(name string, s *Service) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return k.Set(name, data)
}
