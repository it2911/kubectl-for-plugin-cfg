/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
)

// RenameContextOptions contains the options for running the rename-context cli command.
type RenameContextOptions struct {
	ConfigAccess clientcmd.ConfigAccess
	ContextName  string
	NewName      string
}

const (
	renameContextUse = "rename-context CONTEXT_NAME NEW_NAME"

	renameContextShort = "Renames a context from the kubeconfig file."
)

var (
	renameContextLong = templates.LongDesc(`
		Renames a context from the kubeconfig file.

		CONTEXT_NAME is the context name that you wish to change.

		NEW_NAME is the new name you wish to set.

		Note: In case the context being renamed is the 'current-context', this field will also be updated.`)

	renameContextExample = templates.Examples(`
		# Rename the context 'old-name' to 'new-name' in your kubeconfig file
		kubectl config rename-context old-name new-name`)
)

// NewCmdConfigRenameContext creates a command object for the "rename-context" action
func NewCmdConfigRenameContext(out io.Writer, configAccess clientcmd.ConfigAccess) *cobra.Command {
	options := &RenameContextOptions{ConfigAccess: configAccess}

	cmd := &cobra.Command{
		Use:                   renameContextUse,
		DisableFlagsInUseLine: true,
		Short:                 renameContextShort,
		Long:                  renameContextLong,
		Example:               renameContextExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(options.Complete(cmd, args, out))
			cmdutil.CheckErr(options.Validate())
			cmdutil.CheckErr(options.RunRenameContext(out))
		},
	}
	return cmd
}

// Complete assigns RenameContextOptions from the args.
func (o *RenameContextOptions) Complete(cmd *cobra.Command, args []string, out io.Writer) error {
	if len(args) != 2 {
		return helpErrorf(cmd, "Unexpected args: %v", args)
	}

	o.ContextName = args[0]
	o.NewName = args[1]
	return nil
}

// Validate makes sure that provided values for command-line options are valid
func (o RenameContextOptions) Validate() error {
	if len(o.NewName) == 0 {
		return errors.New("You must specify a new non-empty context name")
	}
	return nil
}

// RunRenameContext performs the execution for 'config rename-context' sub command
func (o RenameContextOptions) RunRenameContext(out io.Writer) error {
	config, err := o.ConfigAccess.GetStartingConfig()
	if err != nil {
		return err
	}

	configFile := o.ConfigAccess.GetDefaultFilename()
	if o.ConfigAccess.IsExplicitFile() {
		configFile = o.ConfigAccess.GetExplicitFile()
	}

	context, exists := config.Contexts[o.ContextName]
	if !exists {
		return fmt.Errorf("cannot rename the context %q, it's not in %s", o.ContextName, configFile)
	}

	_, newExists := config.Contexts[o.NewName]
	if newExists {
		return fmt.Errorf("cannot rename the context %q, the context %q already exists in %s", o.ContextName, o.NewName, configFile)
	}

	config.Contexts[o.NewName] = context
	delete(config.Contexts, o.ContextName)

	if config.CurrentContext == o.ContextName {
		config.CurrentContext = o.NewName
	}

	if err := clientcmd.ModifyConfig(o.ConfigAccess, *config, true); err != nil {
		return err
	}

	fmt.Fprintf(out, "Context %q renamed to %q.\n", o.ContextName, o.NewName)
	return nil
}
