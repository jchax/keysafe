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

As a convenience, the input can be decrypted by gpg or clevis tang.`,
		PreRun:                initVars,
		Run:                   restoreRun,
		Args:                  cobra.MaximumNArgs(1),
		DisableFlagsInUseLine: true,
	}
)

func init() {
	rootCmd.AddCommand(restoreCmd)
	f := restoreCmd.Flags()
	f.Bool("gpg", false, "decrypt input using gpg")
	f.Bool("tang", false, "dcrypt input using clevis tang")
	f.BoolVarP(&doClear, "clear", "c", false, "clear keyring before restoring")
	restoreCmd.MarkFlagsMutuallyExclusive("gpg", "tang")
}

func restoreRun(cmd *cobra.Command, args []string) {
	var err error
	var f io.ReadCloser
	if len(args) == 0 {
		f = os.Stdin
	} else {
		f, err = os.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}
	}
	if gpg, _ := cmd.Flags().GetBool("gpg"); gpg {
		cmd := exec.Command("gpg", "--quiet", "--decrypt")
		cmd.Stderr = os.Stderr
		stdin, err := cmd.StdinPipe()
		cobra.CheckErr(err)
		go func(src io.ReadCloser) {
			io.Copy(stdin, src)
			stdin.Close()
			src.Close()
		}(f)
		stdout, err := cmd.StdoutPipe()
		cobra.CheckErr(err)
		if err = cmd.Start(); err != nil {
			log.Fatal(err)
		}
		doRestore(stdout)
		if err = cmd.Wait(); err != nil {
			log.Fatal(err)
		}
		return
	}
	if tang, _ := cmd.Flags().GetBool("tang"); tang {
		cmd := exec.Command("clevis-decrypt-tang")
		cmd.Stderr = os.Stderr
		stdin, err := cmd.StdinPipe()
		cobra.CheckErr(err)
		go func(src io.ReadCloser) {
			io.Copy(stdin, src)
			stdin.Close()
			src.Close()
		}(f)
		stdout, err := cmd.StdoutPipe()
		cobra.CheckErr(err)
		if err = cmd.Start(); err != nil {
			log.Fatal(err)
		}
		doRestore(stdout)
		if err = cmd.Wait(); err != nil {
			log.Fatal(err)
		}
		return
	}
	doRestore(f)
}

func doRestore(src io.ReadCloser) {
	defer src.Close()
	var values map[string]string
	if err := json.NewDecoder(src).Decode(&values); err != nil {
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

func commandPipe(f io.ReadCloser, prog string, args ...string) io.ReadCloser {
	cmd := exec.Command(prog, args...)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		io.Copy(stdin, f)
		stdin.Close()
		//f.Close()
	}()
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return stdout
}
