// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/spf13/cobra"

	"extbldr/impl"
)

// createSrpmCmd represents the createSrpm command
var createSrpmCmd = &cobra.Command{
	Use:   "createSrpm -r <repo> [-p <package>]",
	Short: "Build modified SRPM",
	Long: `A new SRPM is built based on the manifest, spec file and sources specified.
The sources are expected to be already available in <SrcDir>/<repo>.
The manifest might specify only a single package per repo in the general case.
In situations where multiple SRPMs need to be built in dependency order, the manifest might specify multple packages. The [ -p <package> ] can also be used.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")
		pkg, _ := cmd.Flags().GetString("package")
		err := impl.CreateSrpm(repo, pkg)
		return err
	},
}

func init() {
	rootCmd.AddCommand(createSrpmCmd)
}