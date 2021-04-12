// promprobe
// Copyright (C) 2021  Honza Pokorny <honza@pokorny.ca>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/honza/promprobe/pkg/probe"
)

var rootCmd = &cobra.Command{
	Use:   "promprobe",
	Short: "Prometheus Prober",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var memoryCmd = &cobra.Command{
	Use: "memory",
	Run: func(cmd *cobra.Command, args []string) {
		probe.ProbeMemory(cfgFile)
	},
}

var cpuCmd = &cobra.Command{
	Use: "cpu",
	Run: func(cmd *cobra.Command, args []string) {
		probe.ProbeCPU(cfgFile)
	},
}

var cfgFile string

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file ")
	rootCmd.AddCommand(memoryCmd)
	rootCmd.AddCommand(cpuCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
