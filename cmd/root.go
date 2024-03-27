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
	"os"
	"strings"

	"github.com/jchax/keysafe/keyctl"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	keysafe *keyctl.KeySafe
	rootCmd = &cobra.Command{
		Use:   "keysafe",
		Short: "keysafe: keyctl password management",
		Long: `keysafe: keyctl password management.

keysafe manages authentication data on behalf of other applications.
Applications use the keyctl package also defined here.`,
	}
	doJson bool // used by get, set
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.keysafe.yaml)")
	pf.StringP("keyring", "k", keyctl.DefaultKeySafe, "keyring name")
	pf.MarkHidden("config")
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	viper.BindPFlags(pf)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".keysafe")
	}
	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func initLog(cmd *cobra.Command, args []string) {
	log.SetOutput(os.Stderr)
	main, _, _ := strings.Cut(rootCmd.Use, " ")
	sub, _, _ := strings.Cut(cmd.Use, " ")
	log.SetPrefix(main + " " + sub + ": ")
	log.SetFlags(0)
}

func initVars(cmd *cobra.Command, args []string) {
	initLog(cmd, args)
	var err error
	keysafe, err = keyctl.NewKeySafe(viper.GetString("keyring"))
	cobra.CheckErr(err)
}

// This is called before cmd.PreRun (initVars) so keysafe is nil and, of
// course, the program is no longer running once the shell has our
// potential completions.
func singleNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	k, err := keyctl.NewKeySafe(viper.GetString("keyring"))
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	names, err := k.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func multiNameCompletion(cnd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	k, err := keyctl.NewKeySafe(viper.GetString("keyring"))
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	names, err := k.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var res []string
	for _, name := range names {
		var seen bool
		for _, arg := range args {
			if arg == name {
				seen = true
				break
			}
		}
		if !seen {
			res = append(res, name)
		}
	}
	return res, cobra.ShellCompDirectiveNoFileComp
}

func noCompletionArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
