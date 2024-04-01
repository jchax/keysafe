/*
Copyright © 2024 John Haxby <jch@thehaxbys.co.uk>

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

// Encrypt and Decrypt convenience functions.  All three of these functions
// are essentially a wrapper around an io.Reader.  None of them, however,
// are streaming they all consume all of their input before returnng and
// return a complete buffer that can be copied to a destination with no
// further errors.

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"

	"filippo.io/age"
	"filippo.io/age/armor"
)

// encrypt using age (see https://filippo.io/age)
func ageEncrypt(r io.Reader) (io.Reader, error) {
	password, err := getPassword("password", true)
	if err != nil {
		return nil, err
	}
	recipient, err := age.NewScryptRecipient(password)
	if err != nil {
		return nil, err
	}
	var res bytes.Buffer
	w := armor.NewWriter(&res)
	defer w.Close()
	if w, err = age.Encrypt(w, recipient); err != nil {
		return nil, err
	}
	if _, err = io.Copy(w, r); err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	return &res, nil
}

// decrypt using age (see https://filippo.io/age)
func ageDecrypt(r io.Reader) (io.Reader, error) {
	password, err := getPassword("password", false)
	if err != nil {
		log.Fatal(err)
	}
	identity, err := age.NewScryptIdentity(password)
	if err != nil {
		return nil, err
	}
	if r, err = age.Decrypt(armor.NewReader(r), identity); err != nil {
		return nil, err
	}
	var res bytes.Buffer
	if _, err = io.Copy(&res, r); err != nil {
		return nil, err
	}
	return &res, nil
}

// encrypt and decrypt using an external program
func externalCipher(r io.Reader, cmd *exec.Cmd) (io.Reader, error) {
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	defer stdin.Close()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	go func() {
		io.Copy(stdin, r)
		stdin.Close()
	}()
	var res bytes.Buffer
	if _, err = io.Copy(&res, stdout); err != nil {
		return nil, err
	}
	if err = cmd.Wait(); err != nil {
		return nil, err
	}
	return &res, nil
}
