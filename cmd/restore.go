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
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	doClear    bool
	restoreCmd = &cobra.Command{
		Use:   "restore [OPTIONS] [FILENAME]",
		Short: "Restore a keysafe saved by \"keysafe dump\"",
		Long: `Restore a keysafe saved by "keysafe dump".

As a convenience, the input can be decrypted by age, clevis tang or gpg.`,
		PreRun:                initVars,
		Run:                   restoreRun,
		Args:                  cobra.MaximumNArgs(1),
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     singleFileCompletionArgs,
	}
	doAge, doTang, doGpg bool
)

func init() {
	rootCmd.AddCommand(restoreCmd)
	f := restoreCmd.Flags()
	f.BoolVar(&doGpg, "gpg", false, "decrypt input using gpg")
	f.BoolVar(&doTang, "tang", false, "decrypt input using clevis tang")
	f.BoolVar(&doAge, "age", false, "decrypt input using age")
	f.BoolVarP(&doClear, "clear", "c", false, "clear keyring before restoring")
	restoreCmd.MarkFlagsMutuallyExclusive("gpg", "tang", "age")
}

func restoreRun(cmd *cobra.Command, args []string) {
	var r io.Reader = os.Stdin
	if len(args) > 0 {
		in, err := os.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}
		defer in.Close()
		r = in
	}
	var err error
	switch {
	case doGpg:
		r, err = externalCipher(r, exec.Command("gpg", "--quiet", "--decrypt"))
	case doTang:
		r, err = externalCipher(r, exec.Command("clevis-decrypt-tang"))
	case doAge:
		r, err = ageDecrypt(r)
	}
	if err != nil {
		log.Fatalf("decrypt: %v", err)
	}
	var values map[string]string
	if err = json.NewDecoder(r).Decode(&values); err != nil {
		log.Fatalf("json: %v", err)
	}
	if doClear {
		if err := keysafe.Clear(); err != nil {
			log.Fatal(err)
		}
	}
	for name, value := range values {
		if err := keysafe.Set(name, []byte(value)); err != nil {
			log.Fatal(err)
		}
	}
}
