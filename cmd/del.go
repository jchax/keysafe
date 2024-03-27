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
	"log"

	"github.com/spf13/cobra"
)

// delCmd represents the del command
var delCmd = &cobra.Command{
	Use:   "del [OPTIONS] [NAME ...]",
	Short: "Delete entries from the keyring",
	Long: `Delete entries from the keyring.

A keyring itself is never destroyed by the keysafe command but it can be
completely cleared.`,
	PreRun:                initVars,
	Run:                   delRun,
	DisableFlagsInUseLine: true,
}

func init() {
	rootCmd.AddCommand(delCmd)
	f := delCmd.Flags()
	f.BoolVarP(&doClear, "all", "a", false, "remove all names")
	f.BoolVarP(&doClear, "clear", "c", false, "remove all names")
	f.MarkHidden("clear")
}

func delRun(cmd *cobra.Command, args []string) {
	if doClear {
		if err := keysafe.Clear(); err != nil {
			log.Fatal(err)
		}
		// not much point in deleting non-existent entries now
		return
	}
	for _, arg := range args {
		if err := keysafe.Del(arg); err != nil {
			log.Fatal(err)
		}
	}
}
