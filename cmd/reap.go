/*
Copyright © 2026 John Haxby <jch@thehaxbys.co.uk>

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

	"github.com/spf13/cobra"
)

var reapCmd = &cobra.Command{
	Use:   "reap [OPTIONS]",
	Short: "Remove expired keys from a keyring",
	Long: `Removed expired keys froma keyring.

Prints the number of successfully removed keys.`,
	PreRun:                initVars,
	Run:                   reapRun,
	Args:                  cobra.ExactArgs(0),
	DisableFlagsInUseLine: true,
	ValidArgsFunction:     noCompletionArgs,
}

func init() {
	rootCmd.AddCommand(reapCmd)
}

func reapRun(cmd *cobra.Command, args []string) {
	count, err := keysafe.Reap()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d keys reaped\n", count)
}
