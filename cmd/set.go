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
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	setCmd = &cobra.Command{
		Use:   "set [OPTIONS] NAME KEY[+INDICATOR]=VALUE [...]",
		Short: "Set values in an entry",
		Long: `Set values in an entry.

Entries are stored as a JSON dictionary with the specified KEYs and VALUEs.

A +INDICATOR on the end of a modifies the value stored:

    ENV    use the contents of the environment variable VALUE
    FILE   use the contents of the file VALUE, up to 1024 bytes,
           with leading and trailing space stripped
    PROMPT use VALUE as a prompt for a value read as a password`,
		PreRun:                initVars,
		Args:                  cobra.MinimumNArgs(1),
		Run:                   setRun,
		DisableFlagsInUseLine: true,
	}
	timeout time.Duration
)

func init() {
	rootCmd.AddCommand(setCmd)
	f := setCmd.Flags()
	f.DurationVarP(&timeout, "timeout", "t", 0, "set a lifetime for NAME")
	f.BoolVarP(&doJson, "json", "j", false, "set additional values from JSON read from stdin")
}

func setRun(cmd *cobra.Command, args []string) {
	values := make(map[string]string)
	if doJson {
		if err := json.NewDecoder(os.Stdin).Decode(&values); err != nil {
			log.Fatal(err)
		}
	} else {
		cobra.CheckErr(cobra.MinimumNArgs(2)(cmd, args))
	}
	for _, arg := range args[1:] {
		tag, val, err := getTagValue(arg)
		if err != nil {
			log.Fatal(err)
		}
		values[tag] = val
	}
	data, err := json.Marshal(values)
	if err != nil {
		log.Fatal(err)
	}
	if err := keysafe.Set(args[0], data); err != nil {
		log.Fatal(err)
	}
}

func getTagValue(arg string) (string, string, error) {
	tag, val, ok := strings.Cut(arg, "=")
	if !ok {
		return arg, "", nil
	}
	switch {
	case strings.HasSuffix(tag, "+ENV"):
		tag = tag[:len(tag)-4]
		val = os.Getenv(val)
	case strings.HasSuffix(tag, "+FILE"):
		tag = tag[:len(tag)-5]
		data, err := os.ReadFile(val)
		if err != nil {
			return "", "", err
		}
		val = string(bytes.TrimSpace(data))
	case strings.HasSuffix(tag, "+PROMPT"):
		tag = tag[:len(tag)-7]
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			return "", "", err
		}
		defer tty.Close()
		fmt.Fprintf(tty, "%s: ", val)
		data, err := term.ReadPassword(int(tty.Fd()))
		if err != nil {
			return "", "", err
		}
		fmt.Fprintln(tty)
		val = string(data)
	}
	if tag == "" {
		return "", "", errors.New("missing tag")
	}
	if len(val) > 1024 {
		return "", "", errors.New("value too big")
	}
	return tag, val, nil
}
