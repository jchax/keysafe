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
	"fmt"
	"log"
	"sort"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:                   "list [OPTIONS]",
	Short:                 "List known keys",
	Long:                  `List known keys.`,
	PreRun:                initVars,
	Run:                   listRun,
	Args:                  cobra.ExactArgs(0),
	DisableFlagsInUseLine: true,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listRun(cmd *cobra.Command, args []string) {
	names, err := keysafe.List()
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
}
