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

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// dumpCmd represents the dump command
var (
	dumpCmd = &cobra.Command{
		Use:   "dump [FILE]",
		Short: "Dump all keys as JSON",
		Long: `Dump all keys as JSON.

A keysafe can be restored using "keysafe restore".  As a convenience, the
dumped keysafe can be stored in an encrypted file to save it across a
reboot or transferred to another system using ssh.

If the output FILE is given, an existing file is renamed before writing to
the new file.`,
		PreRun:                initVars,
		Run:                   dumpRun,
		Args:                  cobra.MaximumNArgs(1),
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     singleFileCompletionArgs,
	}
	clevisTangUrl   string
	gpgEncryptRecip string
	agePassword     bool
)

func init() {
	rootCmd.AddCommand(dumpCmd)
	f := dumpCmd.Flags()
	f.StringVar(&clevisTangUrl, "tang", "", "encrypt using this clevis tang URL")
	f.StringVar(&gpgEncryptRecip, "gpg", "", "encrypt using gpg for this recipient")
	f.BoolVar(&agePassword, "age", false, "encrypt using age with prompted password")
	dumpCmd.MarkFlagsMutuallyExclusive("tang", "gpg", "age")
}

func dumpRun(cmd *cobra.Command, args []string) {
	names, err := keysafe.List()
	if err != nil {
		log.Fatal(err)
	}
	dump := make(map[string]string)
	for _, name := range names {
		val, err := keysafe.Get(name)
		if err != nil {
			log.Fatal(err)
		}
		dump[name] = string(val)
	}
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(dump); err != nil {
		log.Fatalf("json: %v", err)
	}
	var r io.Reader = &buf

	switch {
	case clevisTangUrl != "":
		r, err = externalCipher(r,
			exec.Command("clevis-encrypt-tang", fmt.Sprintf(`{"url":%q}`, clevisTangUrl), "-y"))
	case gpgEncryptRecip != "":
		r, err = externalCipher(r,
			exec.Command("gpg", "--encrypt", "--armour", "--recipient", gpgEncryptRecip))
	case agePassword:
		r, err = ageEncrypt(r)
	}
	if err != nil {
		log.Fatalf("encrypt: %v", err)
	}
	var w io.Writer = os.Stdout
	if len(args) > 0 {
		err = os.Rename(args[0], args[0]+"~")
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Fatal(err)
		}
		dst, err := os.Create(args[0])
		if err != nil {
			log.Fatal(err)
		}
		defer dst.Close()
		w = dst
	}
	if _, err = io.Copy(w, r); err != nil {
		log.Fatal(err)
	}
}
