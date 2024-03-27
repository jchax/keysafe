//go:build linux

/* A simple keysafe using the Linux kernel key management facility.

A keysafe instance behaves like a map but it backed by the kernel's
persistent keyring.  A new KeySafe instance either creates a named
keyring or retrieves an existing one.  Keys and values are strings.
For example:

    s := keyctl.NewKeySafe("deposit-box")
    s.Set("spare-key", "under the doormat")
    s.Set("combination", "271828")
    for _, k := range s.List() {
        fmt.Println(k, s.Get(k))
    }

The instance's keyring is remove from the persistent keyring with
Unlink().  If the keyring is not linked to any other keyring it will
become inaccessible and eventually be garbage collected by the kernel.

To see the keyring with keyctl(1):

    $ keyctl get_persistent @s
    $ keyctl show

The name of the keyring will be visible under the persistent keyring
that was linked to the session key.q
*/

package keyctl

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

const (
	DefaultKeySafe = "keysafe"
	ErrNoKey       = unix.ENOKEY
	ErrKeyExpired  = unix.EKEYEXPIRED
)

var (
	safeM sync.Mutex
	safes = map[string]*KeySafe{}
)

type KeySafe struct {
	m   sync.Mutex
	err error
	req chan *request
	res chan *response
}

// generic request and response
type request struct {
	op interface{}
}
type response struct {
	err  error
	fin  bool
	buf  []byte
	strs []string
}

// request types
type unlink struct{}
type set struct {
	name  string
	value []byte
}
type setTimeout struct {
	name    string
	timeout time.Duration
}
type get string
type del string
type list struct{}
type clear struct{}

// Create or retrive a keysafe called @name
func NewKeySafe(kname string) (*KeySafe, error) {
	safeM.Lock()
	defer safeM.Unlock()
	if k, ok := safes[kname]; ok {
		return k, nil
	}
	req := make(chan *request)
	res := make(chan *response)
	ch := make(chan error)
	go func(name string) {
		runtime.LockOSThread()
		persistent_keyring, err := unix.KeyctlInt(
			unix.KEYCTL_GET_PERSISTENT, -1,
			unix.KEY_SPEC_PROCESS_KEYRING, 0, 0)
		if err != nil {
			ch <- err
			return
		}
		keyring, err := unix.KeyctlSearch(persistent_keyring, "keyring", name, -1)
		if err == unix.ENOKEY {
			keyring, err = unix.AddKey("keyring", name, nil, persistent_keyring)
		}
		if err != nil {
			ch <- err
			return
		}
		// All OK, we're ready to go now
		close(ch)
		for {
			r := <-req
			switch r.op.(type) {
			case unlink:
				err := doUnlink(persistent_keyring, keyring)
				res <- &response{err: err, fin: true}
				safeM.Lock()
				delete(safes, name)
				safeM.Unlock()
				return
			case set:
				err := doSet(keyring, r.op.(set))
				res <- &response{err: err}
			case setTimeout:
				err := doSetTimeout(keyring, r.op.(setTimeout))
				res <- &response{err: err}
			case get:
				buf, err := doGet(keyring, r.op.(get))
				res <- &response{err: err, buf: buf}
			case del:
				err := doDel(keyring, r.op.(del))
				res <- &response{err: err}
			case list:
				l, err := doList(keyring)
				res <- &response{err: err, strs: l}
			case clear:
				err := doClear(keyring)
				res <- &response{err: err}
			default:
				res <- &response{err: fmt.Errorf("unimplemented operation")}
			}
		}
	}(kname)
	if err, ok := <-ch; ok {
		close(req)
		close(res)
		return nil, err
	}
	safes[kname] = &KeySafe{
		req: req,
		res: res}
	return safes[kname], nil
}

func (k *KeySafe) request(req interface{}) *response {
	k.m.Lock()
	defer k.m.Unlock()
	if k.err != nil {
		return &response{err: k.err}
	}
	k.req <- &request{req}
	resp := <-k.res
	if resp.fin {
		close(k.req)
		close(k.res)
		k.err = ErrNoKey
	}
	return resp
}

func (k *KeySafe) Unlink() error {
	res := k.request(unlink{})
	return res.err
}

func doUnlink(persistent_keyring, keyring int) error {
	_, err := unix.KeyctlInt(unix.KEYCTL_UNLINK, keyring, persistent_keyring, 0, 0)
	return err
}

// Store @value under @name in the keyring.
func (k *KeySafe) Set(name string, value []byte) error {
	res := k.request(set{name, value})
	return res.err
}

func doSet(keyring int, s set) error {
	_, err := unix.AddKey("user", s.name, s.value, keyring)
	return err
}

// Set an expiry timeout for @name in the keyring
// The timeout is rounded to the nearest second
func (k *KeySafe) SetTimeout(name string, timeout time.Duration) error {
	res := k.request(setTimeout{name, timeout})
	return res.err
}

func doSetTimeout(keyring int, st setTimeout) error {
	id, err := unix.KeyctlSearch(keyring, "user", st.name, -1)
	if err != nil {
		return err
	}
	t := int(st.timeout.Round(time.Second).Seconds())
	_, err = unix.KeyctlInt(unix.KEYCTL_SET_TIMEOUT, id, t, 0, 0)
	return err
}

// Locate and retrieve the value associated with @name.
//
// If there is no value associated with @name, the error is usually
// unix.ENOKEY.  See keyctl(2) for possible errors.
func (k *KeySafe) Get(name string) ([]byte, error) {
	res := k.request(get(name))
	return res.buf, res.err
}

func doGet(keyring int, name get) ([]byte, error) {
	id, err := unix.KeyctlSearch(keyring, "user", string(name), -1)
	if err != nil {
		return nil, err
	}
	len, err := unix.KeyctlBuffer(unix.KEYCTL_READ, id, nil, 0)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, len)
	_, err = unix.KeyctlBuffer(unix.KEYCTL_READ, id, buf, 0)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// Remove the entry, if any, under @name.
func (k *KeySafe) Del(name string) error {
	res := k.request(del(name))
	return res.err
}

func doDel(keyring int, name del) error {
	id, err := unix.KeyctlSearch(keyring, "user", string(name), -1)
	if err != nil {
		if err == unix.ENOKEY {
			return nil
		}
		return err
	}
	_, err = unix.KeyctlInt(unix.KEYCTL_UNLINK, id, keyring, 0, 0)
	return err
}

// Get the list of names stored in the keyring.
func (k *KeySafe) List() ([]string, error) {
	res := k.request(list{})
	return res.strs, res.err
}

func doList(keyring int) ([]string, error) {
	len, err := unix.KeyctlBuffer(unix.KEYCTL_READ, keyring, nil, 0)
	if err != nil {
		return nil, err
	}
	data := make([]byte, len)
	_, err = unix.KeyctlBuffer(unix.KEYCTL_READ, keyring, data, len)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len/4)
	for i := 0; i < len/4; i++ {
		id := binary.LittleEndian.Uint32(data[4*i : 4*(i+1)])
		desc, err := unix.KeyctlString(unix.KEYCTL_DESCRIBE, int(id))
		if err == ErrKeyExpired {
			continue
		}
		if err != nil {
			return nil, err
		}
		names = append(names, strings.SplitN(desc, ";", 5)[4])
	}
	return names, nil
}

// Clear all entries from the keyring
func (k *KeySafe) Clear() error {
	return k.request(clear{}).err
}

func doClear(keyring int) error {
	_, err := unix.KeyctlInt(unix.KEYCTL_CLEAR, keyring, 0, 0, 0)
	return err
}
