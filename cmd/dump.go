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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// dumpCmd represents the dump command
var (
	dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump all keys as JSON",
		Long: `Dump all keys as JSON.

A keysafe can be restored using "keysafe restore".  As a convenience, the
dumped keysafe can be stored in an encrypted file to save it across a
reboot or transferred to another system using ssh.`,
		PreRun:                initVars,
		Run:                   dumpRun,
		DisableFlagsInUseLine: true,
	}
	clevisTangUrl   string
	gpgEncryptRecip string
)

func init() {
	rootCmd.AddCommand(dumpCmd)
	f := dumpCmd.Flags()
	f.StringVar(&clevisTangUrl, "tang", "", "encrypt using this clevis tang URL")
	f.StringVar(&gpgEncryptRecip, "gpg", "", "encrypt using gpg for this recipient")
	dumpCmd.MarkFlagsMutuallyExclusive("tang", "gpg")
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
	switch {
	case clevisTangUrl != "":
		cmd := exec.Command("clevis-encrypt-tang",
			fmt.Sprintf(`{"url":%q}`, clevisTangUrl), "-y")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		go func() {
			defer stdin.Close()
			err := json.NewEncoder(stdin).Encode(dump)
			if err != nil {
				log.Fatal(err)
			}
		}()
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	case gpgEncryptRecip != "":
		cmd := exec.Command("gpg", "--encrypt", "--armour", "--recipient", gpgEncryptRecip)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		go func() {
			defer stdin.Close()
			err := json.NewEncoder(stdin).Encode(dump)
			if err != nil {
				log.Fatal(err)
			}
		}()
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	default:
		err = json.NewEncoder(os.Stdout).Encode(dump)
		if err != nil {
			log.Fatal(err)
		}
	}
}
