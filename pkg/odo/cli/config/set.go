package config

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/odo/cli/ui"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const setCommandName = "set"

var (
	setLongDesc = ktemplates.LongDesc(`Set an individual value in the Odo configuration file.

%[1]s`)
	setExample = ktemplates.Examples(`
   # Set a configuration value in the local config
   %[1]s %[2]s java
   %[1]s %[3]s test 
   %[1]s %[4]s 50M 
   %[1]s %[5]s 500M
   %[1]s %[6]s 250M
   %[1]s %[7]s false 
   %[1]s %[8]s 0.5 
   %[1]s %[9]s 2 
   %[1]s %[10]s 1 
	`)
)

// SetOptions encapsulates the options for the command
type SetOptions struct {
	paramName       string
	paramValue      string
	configForceFlag bool
}

// NewSetOptions creates a new SetOptions instance
func NewSetOptions() *SetOptions {
	return &SetOptions{}
}

// Complete completes SetOptions after they've been created
func (o *SetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.paramName = args[0]
	o.paramValue = args[1]
	return
}

// Validate validates the SetOptions based on completed values
func (o *SetOptions) Validate() (err error) {
	return
}

// Run contains the logic for the command
func (o *SetOptions) Run() (err error) {

	cfg, err := config.New()

	if err != nil {
		return errors.Wrapf(err, "unable to set configuration")
	}

	if !o.configForceFlag {
		if isSet := cfg.IsSet(o.paramName); isSet {
			if !ui.Proceed(fmt.Sprintf("%v is already set. Do you want to override it in the config", o.paramName)) {
				fmt.Println("Aborted by the user.")
				return nil
			}
		}
	}

	err = cfg.SetConfiguration(strings.ToLower(o.paramName), o.paramValue)
	if err != nil {
		return err
	}

	fmt.Println("Local config was successfully updated.")
	return nil
}

// NewCmdSet implements the config set odo command
func NewCmdSet(name, fullName string) *cobra.Command {
	o := NewSetOptions()
	configurationSetCmd := &cobra.Command{
		Use:   name,
		Short: "Set a value in odo config file",
		Long:  fmt.Sprintf(setLongDesc, config.FormatLocallySupportedParameters()),
		Example: fmt.Sprintf(fmt.Sprint("\n", setExample), fullName, config.Type,
			config.Name, config.MinMemory, config.MaxMemory, config.Memory, config.Ignore, config.MinCPU, config.MaxCPU, config.CPU),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("please provide a parameter name and value")
			} else if len(args) > 2 {
				return fmt.Errorf("only one value per parameter is allowed")
			} else {
				return nil
			}

		}, Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	configurationSetCmd.Flags().BoolVarP(&o.configForceFlag, "force", "f", false, "Don't ask for confirmation, set the config directly")
	return configurationSetCmd
}
