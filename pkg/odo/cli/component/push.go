package component

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/project"
	"github.com/openshift/odo/pkg/util"

	odoutil "github.com/openshift/odo/pkg/odo/util"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

var pushCmdExample = ktemplates.Examples(`  # Push source code to the current component
%[1]s

# Push data to the current component from the original source.
%[1]s

# Push source code in ~/mycode to component called my-component
%[1]s my-component --context ~/mycode
  `)

// PushRecommendedCommandName is the recommended push command name
const PushRecommendedCommandName = "push"

// PushOptions encapsulates options that push command uses
type PushOptions struct {
	ignores          []string
	sourceType       config.SrcType
	sourcePath       string
	localConfig      *config.LocalConfigInfo
	componentContext string
	cmpConfHash      string
	pushConfig       bool
	pushSource       bool
	*genericclioptions.Context
	show bool
}

// NewPushOptions returns new instance of PushOptions
func NewPushOptions() *PushOptions {
	return &PushOptions{
		ignores:     []string{},
		localConfig: &config.LocalConfigInfo{},
		show:        false,
	}
}

// Complete completes push args
func (po *PushOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {

	conf, err := config.NewLocalConfigInfo(po.componentContext, false)
	if err != nil {
		return errors.Wrap(err, "failed to fetch component config")
	}
	po.localConfig = conf

	po.sourceType = po.localConfig.GetSourceType()
	po.sourcePath = po.localConfig.GetSourceLocation()

	cmpName := po.localConfig.GetName()

	if po.sourceType == config.BINARY || po.sourceType == config.LOCAL {
		u, err := url.Parse(po.sourcePath)
		if err != nil {
			return errors.Wrapf(err, "unable to parse source %s from component %s", po.sourcePath, cmpName)
		}

		if u.Scheme != "" && u.Scheme != "file" {
			return fmt.Errorf("Component %s has invalid source path %s", cmpName, u.Scheme)
		}
		po.sourcePath = util.ReadFilePath(u, runtime.GOOS)
	}

	if len(po.ignores) == 0 {
		rules, err := util.GetIgnoreRulesFromDirectory(po.sourcePath)
		if err != nil {
			odoutil.LogErrorAndExit(err, "")
		}
		po.ignores = append(po.ignores, rules...)
	}
	po.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	return
}

// Validate validates the push parameters
func (po *PushOptions) Validate() (err error) {
	if err = component.ValidateComponentCreateRequest(po.Context.Client, po.localConfig.GetComponentSettings(), false); err != nil {
		return err
	}

	isCmpExists, err := component.Exists(po.Context.Client, po.localConfig.GetName(), po.localConfig.GetApplication())
	if err != nil {
		return err
	}

	if !isCmpExists && po.pushSource && !po.pushConfig {
		return fmt.Errorf("Component %s does not exist and hence cannot push only source. Please use `odo push` without any flags or with both `--source` and `--config` flags", po.localConfig.GetName())
	}

	return nil
}

func (po *PushOptions) createCmpIfNotExistsAndApplyCmpConfig(stdout io.Writer) error {
	if !po.pushConfig {
		// Not the case of component creation or updation(with new config)
		// So nothing to do here and hence return from here
		return nil
	}
	cmpName := po.localConfig.GetName()
	prjName := po.localConfig.GetProject()
	appName := po.localConfig.GetApplication()
	cmpType := po.localConfig.GetType()

	isPrjExists, err := project.Exists(po.Context.Client, prjName)
	if err != nil {
		return errors.Wrapf(err, "failed to check if project with name %s exists", prjName)
	}
	if !isPrjExists {
		log.Successf("Creating project %s", prjName)
		err = project.Create(po.Context.Client, prjName, true)
		if err != nil {
			log.Errorf("Failed creating project %s", prjName)
			return errors.Wrapf(
				err,
				"project %s does not exist. Failed creating it.Please try after creating project using `odo project create <project_name>`",
				prjName,
			)
		}
		log.Successf("Successfully created project %s", prjName)
	}

	isCmpExists, err := component.Exists(po.Context.Client, cmpName, appName)
	if err != nil {
		return errors.Wrapf(err, "failed to check if component %s exists or not", cmpName)
	}

	if !isCmpExists {
		log.Successf("Creating %s component with name %s", cmpType, cmpName)
		// Classic case of component creation
		if err = component.CreateComponent(po.Context.Client, *po.localConfig, po.cmpConfHash, po.componentContext, stdout); err != nil {
			log.Errorf(
				"Failed to create component with name %s. Please use `odo config view` to view settings used to create component. Error: %+v",
				cmpName,
				err,
			)
			os.Exit(1)
		}
		log.Successf("Successfully created component %s", cmpName)
	} else {
		log.Successf("Applying component settings to component: %v", cmpName)
		// Apply config
		err = component.ApplyConfig(po.Context.Client, *po.localConfig, po.cmpConfHash, po.componentContext, stdout)
		if err != nil {
			log.Errorf("Failed to update config to component deployed. Error %+v", err)
			os.Exit(1)
		}
		log.Successf("Successfully created component with name: %v", cmpName)
	}
	return nil
}

// Run has the logic to perform the required actions as part of command
func (po *PushOptions) Run() (err error) {
	stdout := color.Output

	cmpName := po.localConfig.GetName()
	appName := po.localConfig.GetApplication()

	po.resolveSrcAndConfigFlags()

	err = po.createCmpIfNotExistsAndApplyCmpConfig(stdout)
	if err != nil {
		return
	}

	if !po.pushSource {
		// If source is not requested for update, return
		return nil
	}

	log.Successf("Pushing changes to component: %v of type %s", cmpName, po.sourceType)
	glog.V(4).Infof("The client is %#v", po.Context.Client)

	switch po.sourceType {
	case config.LOCAL, config.BINARY:

		if po.sourceType == config.LOCAL {
			glog.V(4).Infof("Copying directory %s to pod", po.sourcePath)
			err = component.PushLocal(
				po.Context.Client,
				cmpName,
				appName,
				po.localConfig.GetSourceLocation(),
				os.Stdout,
				[]string{},
				[]string{},
				true,
				util.GetAbsGlobExps(po.sourcePath, po.ignores),
				po.show,
			)
		} else {
			dir := filepath.Dir(po.sourcePath)
			glog.V(4).Infof("Copying file %s to pod", po.sourcePath)
			err = component.PushLocal(
				po.Context.Client,
				cmpName,
				appName,
				dir,
				os.Stdout,
				[]string{po.sourcePath},
				[]string{},
				true,
				util.GetAbsGlobExps(po.sourcePath, po.ignores),
				po.show,
			)
		}
		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("Failed to push component: %v", cmpName))
		}

	case "git":
		err := component.Build(
			po.Context.Client,
			cmpName,
			appName,
			true,
			stdout,
			po.show,
		)
		return errors.Wrapf(err, fmt.Sprintf("failed to push component: %v", cmpName))
	}

	log.Successf("Changes successfully pushed to component: %v", cmpName)

	return
}

func (po *PushOptions) resolveSrcAndConfigFlags() (err error) {
	cmpName := po.localConfig.GetName()
	appName := po.localConfig.GetApplication()
	// If neither config nor source flag is passed, decide smartly and push only that which changed
	if !po.pushConfig && !po.pushSource {

		// Check smartly if config needs to be updated onto component
		cmpConfigNewHash, err := po.localConfig.Hash()
		glog.V(4).Infof("component config hash is %x and client is %+v", cmpConfigNewHash, po.Context)
		if err != nil {
			glog.V(4).Infof("failed to push to component %s. Error %v", cmpName, err)
		}
		po.cmpConfHash = fmt.Sprintf("%x", cmpConfigNewHash)

		po.pushConfig, err = component.IsUpdateComponentConfig(po.Context.Client, cmpConfigNewHash, cmpName, appName)
		if err != nil {
			glog.V(4).Infof("failed to detect smartly whether the config needs to be updated to component %s. Hence, force applying it. Error %v", cmpName, err)
		}

		// Check smartly if source needs to be updated onto component
		po.pushSource, err = component.IsUpdateComponentSrc(
			po.Context.Client,
			po.localConfig.GetSourceType(),
			po.localConfig.GetSourceLocation(),
			po.localConfig.GetName(),
			po.localConfig.GetApplication(),
		)
		if err != nil {
			glog.V(4).Infof("failed to detect smartly whether to push source or not. Hence, pushing the source. Error : %+v", err)
		}
	}

	glog.V(4).Infof("pushConfig: %+v\npushSource: %+v", po.pushConfig, po.pushSource)

	return nil
}

// NewCmdPush implements the push odo command
func NewCmdPush(name, fullName string) *cobra.Command {
	po := NewPushOptions()

	var pushCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [component name]", name),
		Short:   "Push source code to a component",
		Long:    `Push source code to a component.`,
		Example: fmt.Sprintf(pushCmdExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(po, cmd, args)
		},
	}

	pushCmd.Flags().StringVarP(&po.componentContext, "context", "c", "", "Use given context directory as a source for component settings")
	pushCmd.Flags().BoolVar(&po.show, "show-log", false, "If enabled, logs will be shown when built")
	pushCmd.Flags().StringSliceVar(&po.ignores, "ignore", []string{}, "Files or folders to be ignored via glob expressions.")
	pushCmd.Flags().BoolVar(&po.pushConfig, "config", false, "Use config to update only component settings on to deployed component")
	pushCmd.Flags().BoolVar(&po.pushSource, "source", false, "Use source to deploy only latest source on to deployed component")

	// Add a defined annotation in order to appear in the help menu
	pushCmd.Annotations = map[string]string{"command": "component"}
	pushCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(pushCmd, completion.ComponentNameCompletionHandler)

	return pushCmd
}
