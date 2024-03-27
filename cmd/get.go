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
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	getCmd = &cobra.Command{
		Use:   "get [OPTIONS] NAME",
		Short: "Get the contents of a named key",
		Long: `Get the contents of a named key.

There are options suitable for various script types.  Key contents are
assumed to be well-formed JSON.`,
		PreRun:                initVars,
		Args:                  cobra.ExactArgs(1),
		Run:                   getRun,
		DisableFlagsInUseLine: true,
	}
	doShell bool
)

func init() {
	rootCmd.AddCommand(getCmd)
	f := getCmd.Flags()
	f.BoolVarP(&doShell, "eval", "e", false, "output as Bourne shell settings")
	f.BoolVarP(&doJson, "json", "j", false, "output as JSON")
	getCmd.MarkFlagsMutuallyExclusive("eval", "json")
}

func getRun(cmd *cobra.Command, args []string) {
	data, err := keysafe.Get(args[0])
	if err != nil {
		log.Fatalf("%s: %v", args[0], err)
	}
	var value map[string]string
	err = json.Unmarshal(data, &value)
	if err != nil {
		log.Fatal(err)
	}
	switch {
	case doJson:
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		enc.Encode(&value)
	case doShell:
		for k, v := range value {
			fmt.Printf("%s='%s'\n",
				strings.ToUpper(k), strings.ReplaceAll(v, `'`, `\'`))
		}
	default:
		t := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		defer t.Flush()
		for k, v := range value {
			k1 := strings.ToLower(k)
			if strings.Contains(k1, "pass") || strings.Contains(k1, "secret") {
				v = "????????"
			}
			fmt.Fprintf(t, "%s\t%s\n", k, v)
		}
	}
}
